package web

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/justinas/nosurf"

	"sakuragakure/internal/auth"
	"sakuragakure/internal/db"
	"sakuragakure/internal/iuran"
	"sakuragakure/internal/kas"
	"sakuragakure/internal/konten"
	"sakuragakure/internal/laporan"
	"sakuragakure/internal/media"
	"sakuragakure/internal/web/components"
	"sakuragakure/internal/web/format"
	"sakuragakure/internal/web/pages"
)

// Handlers memegang service untuk halaman web.
type Handlers struct {
	kas     *kas.Service
	iuran   *iuran.Service
	media   *media.Service
	konten  *konten.Service
	laporan *laporan.Service
	auth    *auth.Auth
}

func NewHandlers(conn *sql.DB, a *auth.Auth, mediaDir string) *Handlers {
	return &Handlers{
		kas:     kas.New(conn),
		iuran:   iuran.New(conn),
		media:   media.New(conn, mediaDir),
		konten:  konten.New(conn),
		laporan: laporan.New(conn),
		auth:    a,
	}
}

// nav membangun navigasi sesuai peran user yang sedang masuk.
func (h *Handlers) nav(r *http.Request) []components.NavItem {
	publik := []components.NavItem{
		{Key: "beranda", Label: "Beranda", Href: "/"},
		{Key: "kas", Label: "Kas RT", Href: "/kas"},
		{Key: "kegiatan", Label: "Kegiatan", Href: "/kegiatan"},
		{Key: "aturan", Label: "Aturan", Href: "/aturan"},
		{Key: "pengurus", Label: "Pengurus", Href: "/pengurus"},
	}
	if h.auth.UserID(r.Context()) == 0 {
		return publik
	}
	list, _ := h.auth.PeranList(r.Context())
	for _, p := range list {
		switch p.Peran {
		case "admin":
			return []components.NavItem{
				{Key: "beranda", Label: "Beranda", Href: "/"},
				{Key: "kas", Label: "Kas RT", Href: "/kas"},
				{Key: "kelola", Label: "Kelola user", Href: "/admin/user"},
			}
		case "koordinator", "pembantu_koordinator":
			return []components.NavItem{
				{Key: "beranda", Label: "Beranda", Href: "/"},
				{Key: "tarik", Label: "Tarik iuran", Href: "/gang/" + strconv.Itoa(p.Gang) + "/tarik"},
				{Key: "setor", Label: "Setoran", Href: "/gang/" + strconv.Itoa(p.Gang) + "/setor"},
				{Key: "kas", Label: "Kas RT", Href: "/kas"},
			}
		case "bendahara", "ketua", "wakil":
			return []components.NavItem{
				{Key: "beranda", Label: "Beranda", Href: "/"},
				{Key: "kas", Label: "Kas RT", Href: "/kas"},
				{Key: "setoran", Label: "Setoran", Href: "/kelola/setoran"},
				{Key: "mutasi", Label: "Mutasi", Href: "/kelola/mutasi"},
				{Key: "kelola", Label: "Kelola", Href: "/kelola"},
			}
		case "sekretaris":
			return []components.NavItem{
				{Key: "beranda", Label: "Beranda", Href: "/"},
				{Key: "kas", Label: "Kas RT", Href: "/kas"},
				{Key: "posting", Label: "Posting", Href: "/kelola/konten/posting"},
				{Key: "kelola", Label: "Kelola", Href: "/kelola"},
			}
		}
	}
	return []components.NavItem{
		{Key: "rumahku", Label: "Rumahku", Href: "/rumahku"},
		{Key: "kas", Label: "Kas RT", Href: "/kas"},
		{Key: "kegiatan", Label: "Kegiatan", Href: "/kegiatan"},
		{Key: "aturan", Label: "Aturan", Href: "/aturan"},
	}
}

func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	pages.Login(nosurf.Token(r)).Render(r.Context(), w)
}

func (h *Handlers) Beranda(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.Host, "pengurus.") && h.auth.UserID(r.Context()) == 0 {
		http.Redirect(w, r, "/masuk", http.StatusSeeOther)
		return
	}
	ctx := r.Context()
	periode := r.URL.Query().Get("periode")
	if periode == "" {
		periode = time.Now().Format("2006-01")
	}
	saldo, _ := h.kas.SaldoPosis(ctx)
	kelopak, _ := h.kas.KelopakPerGang(ctx, periode)
	diKoordinator, _ := h.kas.Q.NominalDiKoordinator(ctx)
	pages.Beranda(saldo["kas_rt"], saldo["dana_sosial"], saldo["rukem"], diKoordinator, kelopak, periode, h.nav(r), "beranda").Render(ctx, w)
}

func (h *Handlers) Kas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	saldo, _ := h.kas.SaldoPosis(ctx)
	diKoordinator, _ := h.kas.Q.NominalDiKoordinator(ctx)
	grafik, _ := h.kas.Q.GrafikKasBulanan(ctx)
	mutasi, _ := h.kas.Q.MutasiPublik(ctx, 100)
	dansos, _ := h.kas.Q.DansosTahun(ctx, time.Now().Format("2006"))
	pages.Kas(saldo["kas_rt"], saldo["dana_sosial"], saldo["rukem"], diKoordinator,
		format.ChartSVG(grafik), mutasi, dansos.Jumlah, dansos.Total, h.nav(r), "kas").Render(ctx, w)
}

type tarikVM struct {
	iuran               int64
	bloks               []pages.BlokRumah
	held, done, disetor int
}

func (h *Handlers) tarikData(ctx context.Context, gang int, periode string) tarikVM {
	rows, _ := h.iuran.Q.GridGang(ctx, db.GridGangParams{Gang: int64(gang), Periode: periode})
	iuran, _ := h.iuran.NominalIuran(ctx)
	var vm tarikVM
	vm.iuran = iuran
	byBlok := map[string]*pages.BlokRumah{}
	var order []string
	for _, r := range rows {
		b, ok := byBlok[r.Blok]
		if !ok {
			b = &pages.BlokRumah{Blok: r.Blok}
			byBlok[r.Blok] = b
			order = append(order, r.Blok)
		}
		b.Rumah = append(b.Rumah, r)
		switch r.BayarStatus {
		case "dipegang":
			vm.held++
		case "diterima":
			vm.done++
		case "disetor":
			vm.disetor++
		}
	}
	for _, blok := range order {
		vm.bloks = append(vm.bloks, *byBlok[blok])
	}
	return vm
}

func (h *Handlers) Tarik(w http.ResponseWriter, r *http.Request) {
	gang, _ := strconv.Atoi(chi.URLParam(r, "gang"))
	ctx := r.Context()
	if !h.auth.PunyaGang(ctx, gang, "koordinator", "pembantu_koordinator") {
		http.Error(w, "tidak berhak", http.StatusForbidden)
		return
	}
	periode := periodeOrNow(r)
	d := h.tarikData(ctx, gang, periode)
	pages.Tarik(gang, periode, d.iuran, d.bloks, d.held, d.done, d.disetor, false, nosurf.Token(r), h.nav(r), "tarik").Render(ctx, w)
}

func (h *Handlers) TarikToggle(w http.ResponseWriter, r *http.Request) {
	gang, _ := strconv.Atoi(chi.URLParam(r, "gang"))
	ctx := r.Context()
	if !h.auth.PunyaGang(ctx, gang, "koordinator", "pembantu_koordinator") {
		http.Error(w, "tidak berhak", http.StatusForbidden)
		return
	}
	alamat := r.FormValue("alamat")
	periode := r.FormValue("periode")
	if _, err := h.iuran.TogglePembayaran(ctx, h.auth.UserID(ctx), alamat, periode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	d := h.tarikData(ctx, gang, periode)
	pages.TarikContent(gang, periode, d.iuran, d.bloks, d.held, d.done, d.disetor, false, nosurf.Token(r)).Render(ctx, w)
}

func (h *Handlers) SetorPage(w http.ResponseWriter, r *http.Request) {
	gang, _ := strconv.Atoi(chi.URLParam(r, "gang"))
	ctx := r.Context()
	if !h.auth.PunyaGang(ctx, gang, "koordinator", "pembantu_koordinator") {
		http.Error(w, "tidak berhak", http.StatusForbidden)
		return
	}
	held, _ := h.iuran.Q.HeldPerPeriode(ctx, int64(gang))
	riwayat, _ := h.iuran.Q.SetoranRiwayat(ctx, int64(gang))
	pages.Setor(gang, held, riwayat, nosurf.Token(r), h.nav(r), "setor").Render(ctx, w)
}

func (h *Handlers) SetorSubmit(w http.ResponseWriter, r *http.Request) {
	gang, _ := strconv.Atoi(chi.URLParam(r, "gang"))
	ctx := r.Context()
	if !h.auth.PunyaGang(ctx, gang, "koordinator", "pembantu_koordinator") {
		http.Error(w, "tidak berhak", http.StatusForbidden)
		return
	}
	periode := r.FormValue("periode")
	if _, _, err := h.iuran.Setor(ctx, h.auth.UserID(ctx), gang, periode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/gang/"+strconv.Itoa(gang)+"/setor", http.StatusSeeOther)
}

func (h *Handlers) KelolaSetoran(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pending, _ := h.iuran.Q.SetoranMenunggu(ctx)
	bisaTulis := h.auth.PunyaPeran(ctx, "bendahara", "ketua", "wakil")
	pages.KelolaSetoran(pending, bisaTulis, nosurf.Token(r), h.nav(r), "setoran").Render(ctx, w)
}

func (h *Handlers) TerimaSetoran(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err := h.iuran.TerimaSetoran(ctx, h.auth.UserID(ctx), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/kelola/setoran", http.StatusSeeOther)
}

func (h *Handlers) TolakSetoran(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err := h.iuran.TolakSetoran(ctx, h.auth.UserID(ctx), id, r.FormValue("catatan")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/kelola/setoran", http.StatusSeeOther)
}

func (h *Handlers) MutasiPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	mutasi, _ := h.kas.Q.MutasiSemua(ctx, 200)
	bisaTulis := h.auth.PunyaPeran(ctx, "bendahara", "ketua", "wakil")
	pages.Mutasi(mutasi, bisaTulis, h.nav(r), "mutasi").Render(ctx, w)
}

func (h *Handlers) MutasiBaruPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pos, _ := h.kas.Q.PosDanaList(ctx)
	pages.MutasiBaru(pos, time.Now().Format("2006-01-02"), nosurf.Token(r), h.nav(r), "mutasi").Render(ctx, w)
}

func (h *Handlers) MutasiBaruSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "gagal membaca formulir", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("nota")
	if err != nil {
		http.Error(w, "foto nota wajib dilampirkan", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "gagal membaca foto", http.StatusBadRequest)
		return
	}

	mediaID, err := h.media.Upload(ctx, data, "warga", h.auth.UserID(ctx))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	nominal, _ := strconv.ParseInt(r.FormValue("nominal"), 10, 64)
	keperluan := r.FormValue("keperluan")
	publik := r.FormValue("keterangan_publik")
	if publik == "" {
		publik = keperluan
	}
	if _, err := h.kas.CatatPengeluaran(ctx, h.auth.UserID(ctx), r.FormValue("pos"), r.FormValue("tanggal"),
		r.FormValue("kategori"), keperluan, publik, nominal, mediaID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/kelola/mutasi", http.StatusSeeOther)
}

func (h *Handlers) MediaServe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	m, err := h.media.Get(ctx, id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "kesalahan internal", http.StatusInternalServerError)
		return
	}
	if !h.mediaBoleh(ctx, m) {
		if m.Akses == "warga" {
			http.Redirect(w, r, "/masuk", http.StatusSeeOther)
			return
		}
		http.Error(w, "tidak berhak", http.StatusForbidden)
		return
	}
	path := m.Path
	if chi.URLParam(r, "kind") == "thumb" && m.ThumbPath.Valid {
		path = m.ThumbPath.String
	}
	w.Header().Set("Content-Type", m.Mime)
	http.ServeFile(w, r, path)
}

func (h *Handlers) mediaBoleh(ctx context.Context, m db.Medium) bool {
	switch m.Akses {
	case "publik":
		return true
	case "warga":
		return h.auth.UserID(ctx) != 0
	case "pengurus":
		return h.auth.PunyaPeran(ctx, "ketua", "wakil", "sekretaris", "bendahara")
	case "pemilik":
		return h.auth.UserID(ctx) == m.PemilikUserID.Int64
	}
	return false
}

func (h *Handlers) Aturan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	semua, _ := h.konten.Q.ListAturan(ctx)
	if q == "" {
		pages.Aturan(semua, q, h.nav(r), "aturan").Render(ctx, w)
		return
	}
	var hasil []db.AturanPasal
	for _, p := range semua {
		if strings.Contains(strings.ToLower(p.Judul+" "+p.IsiMd), q) {
			hasil = append(hasil, p)
		}
	}
	pages.Aturan(hasil, q, h.nav(r), "aturan").Render(ctx, w)
}

func (h *Handlers) Pengurus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	daftar, _ := h.konten.Q.ListPengurus(ctx)
	pages.Pengurus(pages.BuildPohon(daftar), h.nav(r), "pengurus").Render(ctx, w)
}

func (h *Handlers) Rumahku(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rumah, err := h.konten.Q.RumahUser(ctx, h.auth.UserID(ctx))
	if err == sql.ErrNoRows {
		pages.Rumahku(db.RumahUserRow{}, false, nil, h.nav(r), "rumahku").Render(ctx, w)
		return
	}
	iuran, _ := h.konten.Q.IuranRumah(ctx, rumah.ID)
	pages.Rumahku(rumah, true, iuran, h.nav(r), "rumahku").Render(ctx, w)
}

func (h *Handlers) Kegiatan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	albums, _ := h.konten.Q.ListAlbum(ctx)
	pages.Kegiatan(albums, h.nav(r), "kegiatan").Render(ctx, w)
}

func (h *Handlers) AlbumDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	a, err := h.konten.Q.AlbumBySlug(ctx, chi.URLParam(r, "slug"))
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	foto, _ := h.konten.Q.FotoAlbum(ctx, a.ID)
	pages.AlbumDetail(a, foto, h.nav(r), "kegiatan").Render(ctx, w)
}

func (h *Handlers) PostingPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pages.Posting(nosurf.Token(r), time.Now().Format("2006-01-02"), h.nav(r), "posting").Render(ctx, w)
}

func (h *Handlers) PostingSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := r.ParseMultipartForm(40 << 20); err != nil {
		http.Error(w, "gagal membaca formulir", http.StatusBadRequest)
		return
	}
	files := r.MultipartForm.File["foto"]
	if len(files) == 0 {
		http.Error(w, "pilih minimal satu foto", http.StatusBadRequest)
		return
	}
	var mediaIDs []int64
	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			http.Error(w, "gagal membaca foto", http.StatusBadRequest)
			return
		}
		data, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			http.Error(w, "gagal membaca foto", http.StatusBadRequest)
			return
		}
		id, err := h.media.Upload(ctx, data, "publik", h.auth.UserID(ctx))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mediaIDs = append(mediaIDs, id)
	}
	if _, err := h.konten.BuatAlbum(ctx, h.auth.UserID(ctx), r.FormValue("judul"), r.FormValue("tanggal"), r.FormValue("cerita"), mediaIDs); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/kegiatan", http.StatusSeeOther)
}

func (h *Handlers) KelolaPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var nama string
	_ = h.auth.DB.QueryRowContext(ctx, `SELECT nama FROM user WHERE id = ?`, h.auth.UserID(ctx)).Scan(&nama)
	pages.Kelola(nama, h.kartuKelola(ctx), h.nav(r), "kelola").Render(ctx, w)
}

func (h *Handlers) kartuKelola(ctx context.Context) []pages.ActionCard {
	list, _ := h.auth.PeranList(ctx)
	var out []pages.ActionCard
	seen := map[string]bool{}
	tambah := func(label, href string) {
		if !seen[href] {
			seen[href] = true
			out = append(out, pages.ActionCard{Label: label, Href: href})
		}
	}

	// baca (transparansi) — semua pengurus bisa lihat
	if h.auth.PunyaPeran(ctx, "ketua", "wakil", "sekretaris", "bendahara", "koordinator", "pembantu_koordinator") {
		tambah("Lihat kas RT (grafik & rangkuman)", "/kas")
		tambah("Laporan & ekspor", "/kelola/laporan")
		tambah("Mutasi kas", "/kelola/mutasi")
		tambah("Setoran", "/kelola/setoran")
	}
	// tulis — sesuai peran
	for _, p := range list {
		switch p.Peran {
		case "bendahara", "wakil":
			tambah("Catat pengeluaran", "/kelola/mutasi/baru")
		case "ketua":
			tambah("Catat pengeluaran", "/kelola/mutasi/baru")
			tambah("Posting kegiatan", "/kelola/konten/posting")
		case "sekretaris":
			tambah("Posting kegiatan", "/kelola/konten/posting")
		case "koordinator", "pembantu_koordinator":
			tambah("Tarik iuran", "/gang/"+strconv.Itoa(p.Gang)+"/tarik")
			tambah("Setoran gang", "/gang/"+strconv.Itoa(p.Gang)+"/setor")
		case "admin":
			tambah("Kelola user", "/admin/user")
		}
	}
	return out
}

func (h *Handlers) AdminUserList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, _ := h.konten.Q.ListUser(ctx)
	pages.AdminUser(users, nosurf.Token(r), h.nav(r), "admin").Render(ctx, w)
}

func (h *Handlers) AdminUserEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	u, err := h.konten.Q.UserByID(ctx, id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	pages.AdminUserEdit(u, nosurf.Token(r), h.nav(r), "admin").Render(ctx, w)
}

func (h *Handlers) AdminUserSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	err := h.konten.Q.UpdateUser(ctx, db.UpdateUserParams{
		Nama:   r.FormValue("nama"),
		NoWa:   auth.NormalisasiWA(r.FormValue("no_wa")),
		Alamat: nullableStr(r.FormValue("alamat")),
		Detail: nullableStr(r.FormValue("detail")),
		ID:     id,
	})
	if err != nil {
		http.Error(w, "gagal menyimpan", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/admin/user", http.StatusSeeOther)
}

func (h *Handlers) LaporanPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data, _ := h.laporan.Data(ctx)
	pages.Laporan(len(data), h.nav(r), "laporan").Render(ctx, w)
}

func (h *Handlers) LaporanCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data, _ := h.laporan.Data(ctx)
	b, _ := laporan.CSV(data)
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="laporan-kas.csv"`)
	w.Write(b)
}

func (h *Handlers) LaporanXLSX(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data, _ := h.laporan.Data(ctx)
	b, err := laporan.XLSX(data)
	if err != nil {
		http.Error(w, "gagal membuat xlsx", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", `attachment; filename="laporan-kas.xlsx"`)
	w.Write(b)
}

func (h *Handlers) LaporanPDF(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data, _ := h.laporan.Data(ctx)
	b, err := laporan.PDF(data)
	if err != nil {
		http.Error(w, "gagal membuat pdf", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="laporan-kas.pdf"`)
	w.Write(b)
}

func nullableStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func periodeOrNow(r *http.Request) string {
	if p := r.URL.Query().Get("periode"); p != "" {
		return p
	}
	return time.Now().Format("2006-01")
}
