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
	"sakuragakure/internal/audit"
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
		case "rw", "perangkat_desa":
			return []components.NavItem{
				{Key: "laporan", Label: "Laporan", Href: "/laporan"},
				{Key: "kas", Label: "Kas RT", Href: "/kas"},
			}
		case "admin":
			return []components.NavItem{
				{Key: "beranda", Label: "Beranda", Href: "/"},
				{Key: "kas", Label: "Kas RT", Href: "/kas"},
				{Key: "admin", Label: "Admin", Href: "/admin"},
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
		{Key: "undangan", Label: "Undangan", Href: "/undangan"},
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
	agenda, _ := h.konten.Q.AgendaMendatang(ctx, db.AgendaMendatangParams{Mulai: time.Now().Format(time.RFC3339), Limit: 4})
	albums, _ := h.konten.Q.ListAlbum(ctx)
	pages.Beranda(saldo["kas_rt"], saldo["dana_sosial"], saldo["rukem"], diKoordinator, kelopak, periode, agenda, albums, h.nav(r), "beranda").Render(ctx, w)
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
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "terima_setoran", "setoran", strconv.FormatInt(id, 10), "", ip(r))
	http.Redirect(w, r, "/kelola/setoran", http.StatusSeeOther)
}

func (h *Handlers) TolakSetoran(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err := h.iuran.TolakSetoran(ctx, h.auth.UserID(ctx), id, r.FormValue("catatan")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "tolak_setoran", "setoran", strconv.FormatInt(id, 10), "", ip(r))
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
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "catat_pengeluaran", "mutasi", "", r.FormValue("pos")+" "+format.FormatRupiah(nominal), ip(r))
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
	pages.Pengurus(pages.BuildPohon(daftar), daftar, h.nav(r), "pengurus").Render(ctx, w)
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
		tambah("Mutasi kas", "/kelola/mutasi")
		tambah("Setoran", "/kelola/setoran")
		tambah("Rukun kematian", "/kelola/rukem")
		tambah("Sinkron data (spreadsheet)", "/kelola/sinkron")
		tambah("Laporan & ekspor", "/kelola/laporan")
	}
	// tulis — sesuai peran
	for _, p := range list {
		switch p.Peran {
		case "bendahara", "wakil":
			tambah("Catat pengeluaran", "/kelola/mutasi/baru")
		case "ketua":
			tambah("Catat pengeluaran", "/kelola/mutasi/baru")
			tambah("Posting kegiatan", "/kelola/konten/posting")
			tambah("Kelola undangan", "/kelola/undangan")
		case "sekretaris":
			tambah("Posting kegiatan", "/kelola/konten/posting")
			tambah("Kelola undangan", "/kelola/undangan")
		case "koordinator", "pembantu_koordinator":
			tambah("Tarik iuran", "/gang/"+strconv.Itoa(p.Gang)+"/tarik")
			tambah("Setoran gang", "/gang/"+strconv.Itoa(p.Gang)+"/setor")
		case "admin":
			tambah("Kelola user", "/admin/user")
		}
	}
	return out
}

func (h *Handlers) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	pages.AdminDashboard(h.nav(r), "admin").Render(r.Context(), w)
}

func (h *Handlers) AdminUserList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, _ := h.konten.Q.ListUser(ctx)
	pages.AdminUser(users, nosurf.Token(r), h.nav(r), "admin").Render(ctx, w)
}

func (h *Handlers) AdminUserBaru(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	peran := r.FormValue("peran")
	gang, _ := strconv.Atoi(r.FormValue("gang"))
	_, err := auth.BuatUser(h.auth.DB, r.FormValue("nama"), r.FormValue("username"), r.FormValue("no_wa"), peran, r.FormValue("password"), gang)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "buat_user", "user", "", r.FormValue("nama")+" ("+peran+")", ip(r))
	http.Redirect(w, r, "/admin/user", http.StatusSeeOther)
}

func (h *Handlers) AdminUserEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	u, err := h.konten.Q.UserByID(ctx, id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	peran, _ := h.konten.Q.UserPeranList(ctx, id)
	pages.AdminUserEdit(u, peran, nosurf.Token(r), h.nav(r), "admin").Render(ctx, w)
}

func (h *Handlers) AdminUserSave(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "gagal membaca formulir", http.StatusBadRequest)
		return
	}
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
	if f, _, ferr := r.FormFile("foto"); ferr == nil {
		data, _ := io.ReadAll(f)
		f.Close()
		if mediaID, uerr := h.media.Upload(ctx, data, "publik", 0); uerr == nil {
			_ = h.konten.Q.UpdateUserFoto(ctx, db.UpdateUserFotoParams{ID: id, FotoMediaID: sql.NullInt64{Int64: mediaID, Valid: true}})
		}
	}
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "ubah_user", "user", strconv.FormatInt(id, 10), "", ip(r))
	http.Redirect(w, r, "/admin/user", http.StatusSeeOther)
}

func (h *Handlers) AdminPeranTambah(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	gang, _ := strconv.Atoi(r.FormValue("gang"))
	_ = h.konten.Q.AddUserPeran(ctx, db.AddUserPeranParams{UserID: id, Peran: r.FormValue("peran"), Gang: int64(gang)})
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "tambah_peran", "user", strconv.FormatInt(id, 10), r.FormValue("peran"), ip(r))
	http.Redirect(w, r, "/admin/user/"+strconv.FormatInt(id, 10), http.StatusSeeOther)
}

func (h *Handlers) AdminPeranHapus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	gang, _ := strconv.Atoi(r.FormValue("gang"))
	_ = h.konten.Q.RemoveUserPeran(ctx, db.RemoveUserPeranParams{UserID: id, Peran: r.FormValue("peran"), Gang: int64(gang)})
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "hapus_peran", "user", strconv.FormatInt(id, 10), r.FormValue("peran"), ip(r))
	http.Redirect(w, r, "/admin/user/"+strconv.FormatInt(id, 10), http.StatusSeeOther)
}

func (h *Handlers) AdminToggleAktif(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	aktif, _ := strconv.Atoi(r.FormValue("aktif"))
	_ = h.konten.Q.ToggleUserAktif(ctx, db.ToggleUserAktifParams{ID: id, Aktif: int64(aktif)})
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "toggle_aktif", "user", strconv.FormatInt(id, 10), "", ip(r))
	http.Redirect(w, r, "/admin/user", http.StatusSeeOther)
}

func (h *Handlers) AdminAudit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	audits, _ := h.konten.Q.ListAudit(ctx)
	pages.AdminAudit(audits, h.nav(r), "admin").Render(ctx, w)
}

func (h *Handlers) AdminGaleri(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	albums, _ := h.konten.Q.ListAlbum(ctx)
	pages.AdminGaleri(albums, nosurf.Token(r), h.nav(r), "admin").Render(ctx, w)
}

func (h *Handlers) AdminGaleriHapus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	_ = h.konten.HapusAlbum(ctx, id)
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "hapus_album", "album", strconv.FormatInt(id, 10), "", ip(r))
	http.Redirect(w, r, "/admin/galeri", http.StatusSeeOther)
}

func (h *Handlers) AdminDokumen(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	media, _ := h.konten.Q.ListMedia(ctx)
	pages.AdminDokumen(media, nosurf.Token(r), h.nav(r), "admin").Render(ctx, w)
}

func (h *Handlers) AdminDokumenHapus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.konten.HapusMedia(ctx, id); err != nil {
		http.Error(w, "media masih dipakai", http.StatusBadRequest)
		return
	}
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "hapus_media", "media", strconv.FormatInt(id, 10), "", ip(r))
	http.Redirect(w, r, "/admin/dokumen", http.StatusSeeOther)
}

func ip(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		return strings.TrimSpace(strings.Split(x, ",")[0])
	}
	return r.RemoteAddr
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

func (h *Handlers) SinkronPage(w http.ResponseWriter, r *http.Request) {
	pages.Sinkron(h.nav(r), "sinkron").Render(r.Context(), w)
}

func (h *Handlers) SinkronCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	table := chi.URLParam(r, "table")
	var header []string
	var rows [][]string

	switch table {
	case "pengurus":
		daftar, _ := h.konten.Q.ListPengurus(ctx)
		header = []string{"Nama", "Jabatan", "Alamat", "WhatsApp"}
		for _, p := range daftar {
			rows = append(rows, []string{p.Nama, pages.PeranLabel(p.Peran, int(p.Gang)), p.Alamat.String, p.NoWa})
		}
	case "mutasi":
		data, _ := h.laporan.Data(ctx)
		header = []string{"Tanggal", "Pos", "Kategori", "Keterangan", "Masuk", "Keluar"}
		for _, b := range data {
			rows = append(rows, []string{b.Tanggal, b.Pos, b.Kategori, b.Keterangan, strconv.FormatInt(b.Masuk, 10), strconv.FormatInt(b.Keluar, 10)})
		}
	case "setoran":
		s, _ := h.konten.Q.ListSetoranSemua(ctx)
		header = []string{"Gang", "Periode", "Total", "Status", "Dibuat", "Koordinator"}
		for _, x := range s {
			rows = append(rows, []string{strconv.FormatInt(x.Gang, 10), x.Periode, strconv.FormatInt(x.Total, 10), x.Status, x.DibuatAt, x.Koordinator})
		}
	case "rumah":
		rm, _ := h.konten.Q.ListRumah(ctx)
		header = []string{"Alamat", "Blok", "Nomor", "Gang", "Status", "Nama KK", "Anggota", "Bebas iuran", "Catatan"}
		for _, x := range rm {
			rows = append(rows, []string{x.Alamat, x.Blok, x.Nomor, strconv.FormatInt(x.Gang, 10), x.Status, x.NamaKk, strconv.FormatInt(x.JumlahAnggota.Int64, 10), x.BebasIuranAlasan.String, x.CatatanValidasi.String})
		}
	default:
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+table+`.csv"`)
	w.Write(format.CSV(header, rows))
}

func (h *Handlers) RukemPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	saldo, _ := h.kas.Q.SaldoPos(ctx, "rukem")
	pages.Rukem(saldo, h.nav(r), "rukem").Render(ctx, w)
}

func (h *Handlers) undanganItems(ctx context.Context, rumahID int64) []pages.UndanganItem {
	list, _ := h.konten.Q.ListUndangan(ctx)
	total, _ := h.konten.Q.WajibIuranRumahCount(ctx)
	items := make([]pages.UndanganItem, 0, len(list))
	for _, u := range list {
		jawaban := ""
		if rumahID != 0 {
			if j, err := h.konten.Q.RsvpJawaban(ctx, db.RsvpJawabanParams{UndanganID: u.ID, RumahID: rumahID}); err == nil {
				jawaban = j
			}
		}
		c, _ := h.konten.Q.RsvpCount(ctx, u.ID)
		items = append(items, pages.UndanganItem{ID: u.ID, Judul: u.Judul, Isi: u.Isi, Terbit: u.TerbitAt, Jawaban: jawaban, Hadir: c.Hadir, Tidak: c.Tidak, Total: total})
	}
	return items
}

func (h *Handlers) Undangan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var rumahID int64
	if rumah, err := h.konten.Q.RumahUser(ctx, h.auth.UserID(ctx)); err == nil {
		rumahID = rumah.ID
	}
	pages.UndanganWarga(h.undanganItems(ctx, rumahID), rumahID != 0, h.nav(r), "undangan").Render(ctx, w)
}

func (h *Handlers) Rsvp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	undanganID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	jawaban := r.FormValue("jawaban")
	if jawaban != "hadir" && jawaban != "tidak" {
		http.Error(w, "jawaban tidak valid", http.StatusBadRequest)
		return
	}
	rumah, err := h.konten.Q.RumahUser(ctx, h.auth.UserID(ctx))
	if err != nil {
		http.Error(w, "akun belum terhubung ke rumah", http.StatusBadRequest)
		return
	}
	_ = h.konten.Q.SetRsvp(ctx, db.SetRsvpParams{UndanganID: undanganID, RumahID: rumah.ID, Jawaban: jawaban, DijawabAt: time.Now().Format(time.RFC3339)})
	http.Redirect(w, r, "/undangan", http.StatusSeeOther)
}

func (h *Handlers) KelolaUndangan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	pages.KelolaUndangan(h.undanganItems(ctx, 0), nosurf.Token(r), h.nav(r), "undangan").Render(ctx, w)
}

func (h *Handlers) BuatUndangan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	judul := r.FormValue("judul")
	isi := r.FormValue("isi")
	if judul == "" || isi == "" {
		http.Error(w, "judul dan isi wajib diisi", http.StatusBadRequest)
		return
	}
	if _, err := h.konten.Q.BuatUndangan(ctx, db.BuatUndanganParams{Judul: judul, Isi: isi, TerbitAt: time.Now().Format(time.RFC3339), DibuatOleh: h.auth.UserID(ctx)}); err != nil {
		http.Error(w, "gagal menerbitkan", http.StatusBadRequest)
		return
	}
	audit.Tulis(h.auth.DB, h.auth.UserID(ctx), "terbit_undangan", "undangan", "", judul, ip(r))
	http.Redirect(w, r, "/kelola/undangan", http.StatusSeeOther)
}

func (h *Handlers) LaporanDesa(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rekap, _ := h.konten.Q.RekapKependudukan(ctx)
	pages.LaporanDesa(rekap, h.nav(r), "laporan").Render(ctx, w)
}

func periodeOrNow(r *http.Request) string {
	if p := r.URL.Query().Get("periode"); p != "" {
		return p
	}
	return time.Now().Format("2006-01")
}
