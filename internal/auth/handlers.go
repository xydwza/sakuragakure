package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/justinas/nosurf"

	"sakuragakure/internal/notif"
)

// ipLimiter membatasi permintaan per IP (10/jam). Cukup untuk satu instance.
type ipLimiter struct {
	mu sync.Mutex
	m  map[string][]time.Time
}

func (l *ipLimiter) Izinkan(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.m == nil {
		l.m = map[string][]time.Time{}
	}
	batas := time.Now().Add(-time.Hour)
	keep := l.m[ip][:0]
	for _, t := range l.m[ip] {
		if t.After(batas) {
			keep = append(keep, t)
		}
	}
	if len(keep) >= 10 {
		l.m[ip] = keep
		return false
	}
	l.m[ip] = append(keep, time.Now())
	return true
}

var limiter = &ipLimiter{}

// LoginPage menampilkan form masuk (nomor WA).
func (a *Auth) LoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Masuk</title></head><body>
<h1>Masuk</h1>
<form method="post" action="/masuk/wa">
  <input type="hidden" name="csrf_token" value="%s">
  <label>Nomor WhatsApp <input name="no_wa" inputmode="tel" placeholder="08xx" required></label>
  <button>Kirim kode</button>
</form>
<hr>
<form method="post" action="/masuk/kode-cadangan">
  <input type="hidden" name="csrf_token" value="%s">
  <label>Kode cadangan <input name="kode" required></label>
  <button>Masuk dengan kode cadangan</button>
</form>
</body></html>`, nosurf.Token(r), nosurf.Token(r))
}

// MintaOTP menangani kirim kode OTP ke nomor yang terdaftar.
func (a *Auth) MintaOTP(w http.ResponseWriter, r *http.Request) {
	noWA := normalisasiWA(r.FormValue("no_wa"))
	if noWA == "" {
		http.Error(w, "nomor WhatsApp kosong", http.StatusBadRequest)
		return
	}
	if !limiter.Izinkan(r.RemoteAddr) {
		http.Error(w, "terlalu banyak permintaan, coba lagi nanti", http.StatusTooManyRequests)
		return
	}

	var userID int64
	err := a.DB.QueryRow(`SELECT id FROM user WHERE no_wa = ? AND aktif = 1 AND (aktif_sampai IS NULL OR aktif_sampai > ?)`,
		noWA, now().Format(time.RFC3339)).Scan(&userID)
	if err != nil {
		http.Error(w, "nomor tidak terdaftar", http.StatusNotFound)
		return
	}

	kode, err := GenerateOTP(a.DB, noWA)
	if err != nil {
		http.Error(w, err.Error(), http.StatusTooManyRequests)
		return
	}
	if err := notif.Enqueue(a.DB, noWA, "otp", kode); err != nil {
		log.Printf("gagal antri OTP: %v", err)
	}
	if os.Getenv("WA_GATEWAY_URL") == "" {
		log.Printf("[dev] kode OTP untuk %s: %s", noWA, kode)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html lang="id"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Kode</title></head><body>
<h1>Masukkan kode</h1>
<p>Kode 6 digit dikirim ke %s.</p>
<form method="post" action="/masuk/kode">
  <input type="hidden" name="csrf_token" value="%s">
  <input type="hidden" name="no_wa" value="%s">
  <input name="kode" inputmode="numeric" maxlength="6" autocomplete="one-time-code" required>
  <button>Masuk</button>
</form>
</body></html>`, htmlEsc(noWA), nosurf.Token(r), htmlEsc(noWA))
}

// VerifikasiOTP mencocokkan kode dan membuka sesi.
func (a *Auth) VerifikasiOTP(w http.ResponseWriter, r *http.Request) {
	noWA := normalisasiWA(r.FormValue("no_wa"))
	kode := strings.TrimSpace(r.FormValue("kode"))
	if err := VerifyOTP(a.DB, noWA, kode); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var userID int64
	if err := a.DB.QueryRow(`SELECT id FROM user WHERE no_wa = ?`, noWA).Scan(&userID); err != nil {
		http.Error(w, "nomor tidak terdaftar", http.StatusNotFound)
		return
	}
	a.masuk(w, r, userID)
}

// KodeCadangan masuk memakai kode sekali pakai.
func (a *Auth) KodeCadangan(w http.ResponseWriter, r *http.Request) {
	kode := strings.TrimSpace(r.FormValue("kode"))
	userID, err := PakaiKodeCadangan(a.DB, kode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	a.masuk(w, r, userID)
}

func (a *Auth) masuk(w http.ResponseWriter, r *http.Request, userID int64) {
	if err := a.Sessions.RenewToken(r.Context()); err != nil {
		http.Error(w, "kesalahan sesi", http.StatusInternalServerError)
		return
	}
	a.Sessions.Put(r.Context(), "userID", userID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Keluar menghancurkan sesi.
func (a *Auth) Keluar(w http.ResponseWriter, r *http.Request) {
	if err := a.Sessions.Destroy(r.Context()); err != nil {
		log.Printf("gagal hapus sesi: %v", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func normalisasiWA(s string) string {
	s = strings.TrimSpace(s)
	s = strings.NewReplacer(" ", "", "-", "").Replace(s)
	if strings.HasPrefix(s, "0") {
		s = "62" + s[1:]
	}
	if !strings.HasPrefix(s, "+") && strings.HasPrefix(s, "62") {
		s = "+" + s
	}
	return s
}

func htmlEsc(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}
