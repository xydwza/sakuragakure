package web

import (
	"database/sql"
	"net/http"
	"time"

	"sakuragakure/internal/kas"
	"sakuragakure/internal/web/format"
	"sakuragakure/internal/web/pages"
)

// Handlers memegang service untuk halaman web.
type Handlers struct {
	kas *kas.Service
}

func NewHandlers(conn *sql.DB) *Handlers {
	return &Handlers{kas: kas.New(conn)}
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
	pages.Beranda(saldo["kas_rt"], saldo["dana_sosial"], saldo["rukem"], diKoordinator, kelopak, periode).Render(ctx, w)
}

func (h *Handlers) Kas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	saldo, _ := h.kas.SaldoPosis(ctx)
	diKoordinator, _ := h.kas.Q.NominalDiKoordinator(ctx)
	grafik, _ := h.kas.Q.GrafikKasBulanan(ctx)
	mutasi, _ := h.kas.Q.MutasiPublik(ctx, 100)
	dansos, _ := h.kas.Q.DansosTahun(ctx, time.Now().Format("2006"))
	pages.Kas(saldo["kas_rt"], saldo["dana_sosial"], saldo["rukem"], diKoordinator,
		format.ChartSVG(grafik), mutasi, dansos.Jumlah, dansos.Total).Render(ctx, w)
}
