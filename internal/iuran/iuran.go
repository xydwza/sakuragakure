package iuran

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"sakuragakure/internal/db"
)

// Service menangani alur iuran: tagihan, pembayaran tunai, setoran, konfirmasi.
type Service struct {
	Q  *db.Queries
	DB *sql.DB
}

func New(conn *sql.DB) *Service { return &Service{Q: db.New(conn), DB: conn} }

func wib() *time.Location { return time.FixedZone("WIB", 7*3600) }
func nowRFC3339() string  { return time.Now().In(wib()).Format(time.RFC3339) }
func todayStr() string    { return time.Now().In(wib()).Format("2006-01-02") }

func (s *Service) NominalIuran(ctx context.Context) (int64, error) {
	var v string
	if err := s.DB.QueryRowContext(ctx, `SELECT value FROM setting WHERE key = 'iuran_kas_bulanan'`).Scan(&v); err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("setting iuran_kas_bulanan tidak valid: %w", err)
	}
	return n, nil
}

// GenerateTagihan membuat tagihan kas untuk semua rumah wajib iuran (idempotent).
func (s *Service) GenerateTagihan(ctx context.Context, periode string) (int64, error) {
	nominal, err := s.NominalIuran(ctx)
	if err != nil {
		return 0, err
	}
	res, err := s.DB.ExecContext(ctx, `INSERT INTO tagihan (rumah_id, penghuni_id, jenis, periode, nominal)
		SELECT r.id, ph.id, 'kas', ?, ?
		FROM rumah r
		LEFT JOIN penghuni ph ON ph.rumah_id = r.id AND ph.selesai IS NULL
		WHERE r.status != 'kosong' AND r.bebas_iuran_alasan IS NULL
		ON CONFLICT(rumah_id, jenis, periode) DO NOTHING`, periode, nominal)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// TogglePembayaran mencatat pembayaran tunai "dipegang" koordinator, atau
// membatalkannya bila masih dipegang. Mengembalikan "dipegang" atau "batal".
func (s *Service) TogglePembayaran(ctx context.Context, koordinatorID int64, alamat, periode string) (string, error) {
	nominal, err := s.NominalIuran(ctx)
	if err != nil {
		return "", err
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var rumahID int64
	var status, bebas string
	err = tx.QueryRowContext(ctx, `SELECT id, status, COALESCE(bebas_iuran_alasan, '') FROM rumah WHERE alamat = ?`, alamat).Scan(&rumahID, &status, &bebas)
	if err == sql.ErrNoRows {
		return "", errors.New("rumah tidak ditemukan")
	}
	if err != nil {
		return "", err
	}
	if status == "kosong" {
		return "", errors.New("rumah kosong, tidak wajib iuran")
	}
	if bebas != "" {
		return "", errors.New("rumah bebas iuran")
	}

	// cari atau buat tagihan
	var tagihanID int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM tagihan WHERE rumah_id = ? AND jenis = 'kas' AND periode = ?`, rumahID, periode).Scan(&tagihanID)
	if err == sql.ErrNoRows {
		var penghuniID sql.NullInt64
		_ = tx.QueryRowContext(ctx, `SELECT id FROM penghuni WHERE rumah_id = ? AND selesai IS NULL`, rumahID).Scan(&penghuniID)
		res, err := tx.ExecContext(ctx, `INSERT INTO tagihan (rumah_id, penghuni_id, jenis, periode, nominal) VALUES (?, ?, 'kas', ?, ?)`,
			rumahID, penghuniID, periode, nominal)
		if err != nil {
			return "", err
		}
		tagihanID, _ = res.LastInsertId()
	} else if err != nil {
		return "", err
	}

	var pbID int64
	var pbStatus string
	err = tx.QueryRowContext(ctx, `SELECT id, status FROM pembayaran WHERE tagihan_id = ?`, tagihanID).Scan(&pbID, &pbStatus)
	if err == sql.ErrNoRows {
		if _, err := tx.ExecContext(ctx, `INSERT INTO pembayaran (tagihan_id, metode, status, dicatat_oleh, dicatat_at) VALUES (?, 'tunai', 'dipegang', ?, ?)`,
			tagihanID, koordinatorID, nowRFC3339()); err != nil {
			return "", err
		}
		return "dipegang", tx.Commit()
	}
	if err != nil {
		return "", err
	}
	if pbStatus == "dipegang" {
		if _, err := tx.ExecContext(ctx, `DELETE FROM pembayaran WHERE id = ?`, pbID); err != nil {
			return "", err
		}
		return "batal", tx.Commit()
	}
	return "", fmt.Errorf("pembayaran sudah %s, tidak bisa diubah", pbStatus)
}

// Setor membuat setoran menunggu untuk gang+periode, mengubah pembayaran
// dipegang menjadi disetor. Mengembalikan id setoran dan total.
func (s *Service) Setor(ctx context.Context, koordinatorID int64, gang int, periode string) (int64, int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	var total int64
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(SUM(t.nominal), 0) FROM pembayaran p
		JOIN tagihan t ON t.id = p.tagihan_id
		JOIN rumah r ON r.id = t.rumah_id
		WHERE p.status = 'dipegang' AND r.gang = ? AND t.periode = ?`, gang, periode).Scan(&total)
	if err != nil {
		return 0, 0, err
	}
	if total == 0 {
		return 0, 0, errors.New("tidak ada iuran yang belum disetor")
	}

	res, err := tx.ExecContext(ctx, `INSERT INTO setoran (gang, periode, koordinator_id, total, status, dibuat_at) VALUES (?, ?, ?, ?, 'menunggu', ?)`,
		gang, periode, koordinatorID, total, nowRFC3339())
	if err != nil {
		return 0, 0, err
	}
	setoranID, _ := res.LastInsertId()

	if _, err := tx.ExecContext(ctx, `UPDATE pembayaran SET status = 'disetor', setoran_id = ?
		WHERE id IN (SELECT p.id FROM pembayaran p JOIN tagihan t ON t.id = p.tagihan_id JOIN rumah r ON r.id = t.rumah_id
			WHERE p.status = 'dipegang' AND r.gang = ? AND t.periode = ?)`, setoranID, gang, periode); err != nil {
		return 0, 0, err
	}
	return setoranID, total, tx.Commit()
}

// TerimaSetoran menerima setoran: pembayaran diterima + satu mutasi masuk kas RT.
func (s *Service) TerimaSetoran(ctx context.Context, bendaharaID, setoranID int64) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	var gang int
	var periode string
	var total int64
	err = tx.QueryRowContext(ctx, `SELECT status, gang, periode, total FROM setoran WHERE id = ?`, setoranID).
		Scan(&status, &gang, &periode, &total)
	if err == sql.ErrNoRows {
		return errors.New("setoran tidak ditemukan")
	}
	if err != nil {
		return err
	}
	if status != "menunggu" {
		return errors.New("setoran sudah diproses")
	}

	if _, err := tx.ExecContext(ctx, `UPDATE setoran SET status = 'diterima', dikonfirmasi_oleh = ?, dikonfirmasi_at = ? WHERE id = ?`,
		bendaharaID, nowRFC3339(), setoranID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE pembayaran SET status = 'diterima' WHERE setoran_id = ?`, setoranID); err != nil {
		return err
	}
	ket := fmt.Sprintf("Setoran iuran gang %d, %s", gang, periode)
	if _, err := tx.ExecContext(ctx, `INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, ref_tipe, ref_id, dibuat_oleh, dibuat_at)
		VALUES ('kas_rt', ?, 'masuk', ?, 'Iuran', ?, ?, 'setoran', ?, ?, ?)`,
		todayStr(), total, ket, ket, setoranID, bendaharaID, nowRFC3339()); err != nil {
		return err
	}
	return tx.Commit()
}

// TolakSetoran menolak setoran: pembayaran kembali ke dipegang.
func (s *Service) TolakSetoran(ctx context.Context, bendaharaID, setoranID int64, catatan string) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	if err := tx.QueryRowContext(ctx, `SELECT status FROM setoran WHERE id = ?`, setoranID).Scan(&status); err != nil {
		return err
	}
	if status != "menunggu" {
		return errors.New("setoran sudah diproses")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE setoran SET status = 'ditolak', dikonfirmasi_oleh = ?, dikonfirmasi_at = ?, catatan = ? WHERE id = ?`,
		bendaharaID, nowRFC3339(), catatan, setoranID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE pembayaran SET status = 'dipegang', setoran_id = NULL WHERE setoran_id = ? AND status = 'disetor'`, setoranID); err != nil {
		return err
	}
	return tx.Commit()
}
