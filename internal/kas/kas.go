package kas

import (
	"context"
	"database/sql"

	"sakuragakure/internal/db"
)

// Kelopak adalah agregat satu gang untuk satu periode iuran.
type Kelopak struct {
	Gang                  int
	Total, Diterima, Dipegang int
}

// Service membaca data kas agregat (publik) dan mencatat mutasi.
type Service struct {
	Q  *db.Queries
	DB *sql.DB
}

func New(conn *sql.DB) *Service { return &Service{Q: db.New(conn), DB: conn} }

// KelopakPerGang menggabungkan total wajib iuran per gang dengan status
// diterima/dipegang untuk satu periode.
func (s *Service) KelopakPerGang(ctx context.Context, periode string) ([]Kelopak, error) {
	totals, err := s.Q.WajibIuranPerGang(ctx)
	if err != nil {
		return nil, err
	}
	stats, err := s.Q.KelopakGang(ctx, periode)
	if err != nil {
		return nil, err
	}

	byGang := map[int64]*Kelopak{}
	for _, t := range totals {
		byGang[t.Gang] = &Kelopak{Gang: int(t.Gang), Total: int(t.Total)}
	}
	for _, st := range stats {
		if k, ok := byGang[st.Gang]; ok {
			k.Diterima = int(st.Diterima)
			k.Dipegang = int(st.Dipegang)
		}
	}
	out := make([]Kelopak, 0, 3)
	for g := int64(1); g <= 3; g++ {
		if k, ok := byGang[g]; ok {
			out = append(out, *k)
		}
	}
	return out, nil
}

// SaldoPosis mengembalikan map pos_id -> saldo.
func (s *Service) SaldoPosis(ctx context.Context) (map[string]int64, error) {
	rows, err := s.Q.SaldoPosis(ctx)
	if err != nil {
		return nil, err
	}
	m := make(map[string]int64, len(rows))
	for _, r := range rows {
		m[r.PosID] = r.Saldo
	}
	return m, nil
}
