package web

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/justinas/nosurf"

	"sakuragakure/internal/auth"
	"sakuragakure/internal/db"
	"sakuragakure/internal/iuran"
	"sakuragakure/internal/kas"
	"sakuragakure/internal/web/components"
	"sakuragakure/internal/web/format"
	"sakuragakure/internal/web/pages"
)

// Handlers memegang service untuk halaman web.
type Handlers struct {
	kas   *kas.Service
	iuran *iuran.Service
	auth  *auth.Auth
}

func NewHandlers(conn *sql.DB, a *auth.Auth) *Handlers {
	return &Handlers{kas: kas.New(conn), iuran: iuran.New(conn), auth: a}
}

// nav membangun navigasi sesuai peran user yang sedang masuk.
func (h *Handlers) nav(r *http.Request) []components.NavItem {
	publik := []components.NavItem{
		{Key: "beranda", Label: "Beranda", Href: "/"},
		{Key: "kas", Label: "Kas RT", Href: "/kas"},
	}
	if h.auth.UserID(r.Context()) == 0 {
		return publik
	}
	list, _ := h.auth.PeranList(r.Context())
	for _, p := range list {
		switch p.Peran {
		case "koordinator", "pembantu_koordinator":
			return []components.NavItem{
				{Key: "tarik", Label: "Tarik iuran", Href: "/gang/" + strconv.Itoa(p.Gang) + "/tarik"},
				{Key: "setor", Label: "Setoran", Href: "/gang/" + strconv.Itoa(p.Gang) + "/setor"},
				{Key: "kas", Label: "Kas RT", Href: "/kas"},
			}
		case "bendahara", "ketua", "wakil", "sekretaris":
			return []components.NavItem{
				{Key: "setoran", Label: "Setoran", Href: "/kelola/setoran"},
				{Key: "kas", Label: "Kas RT", Href: "/kas"},
			}
		}
	}
	return publik
}

func (h *Handlers) Beranda(w http.ResponseWriter, r *http.Request) {
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
	pages.KelolaSetoran(pending, nosurf.Token(r), h.nav(r), "setoran").Render(ctx, w)
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

func periodeOrNow(r *http.Request) string {
	if p := r.URL.Query().Get("periode"); p != "" {
		return p
	}
	return time.Now().Format("2006-01")
}
