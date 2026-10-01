package auth

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"sakuragakure/internal/audit"
)

// Login memverifikasi kata sandi dan menerbitkan token JWT.
func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	input := strings.TrimSpace(r.FormValue("user"))
	pw := r.FormValue("password")

	var id int64
	var hash sql.NullString
	err := a.DB.QueryRow(`SELECT id, password_hash FROM user
		WHERE (LOWER(username) = LOWER(?) OR no_wa = ?) AND aktif = 1 AND (aktif_sampai IS NULL OR aktif_sampai > ?)`,
		input, NormalisasiWA(input), now().Format(time.RFC3339)).Scan(&id, &hash)
	if err != nil || !hash.Valid || !cekPassword(hash.String, pw) {
		http.Error(w, "nama pengguna atau kata sandi salah", http.StatusUnauthorized)
		return
	}

	token, err := a.TerbitkanToken(id)
	if err != nil {
		http.Error(w, "kesalahan sesi", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: tokenCookie, Value: token, Path: "/",
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: cookieSecure(),
		MaxAge: int(sesiTTL.Seconds()),
	})
	audit.Tulis(a.DB, id, "masuk", "sesi", "", "", r.RemoteAddr)
	http.Redirect(w, r, a.tujuanLogin(id), http.StatusSeeOther)
}

// tujuanLogin mengarahkan user sesuai perannya setelah masuk.
func (a *Auth) tujuanLogin(userID int64) string {
	var peran string
	if err := a.DB.QueryRow(`SELECT peran FROM user_peran WHERE user_id = ? ORDER BY urutan LIMIT 1`, userID).Scan(&peran); err != nil {
		return "/"
	}
	switch peran {
	case "admin":
		return "/admin"
	case "warga":
		return "/"
	default:
		return "/kelola"
	}
}

// Logout menghapus cookie token.
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: tokenCookie, Value: "", Path: "/", HttpOnly: true, MaxAge: -1})
	http.Redirect(w, r, "/masuk", http.StatusSeeOther)
}

func NormalisasiWA(s string) string {
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
