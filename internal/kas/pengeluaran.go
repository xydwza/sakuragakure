package kas

import (
	"context"
	"errors"
	"time"
)

// CatatPengeluaran mencatat mutasi keluar bernota. Foto nota wajib (dipaksa
// oleh CHECK di database). Trigger menolak bila periode sudah ditutup.
func (s *Service) CatatPengeluaran(ctx context.Context, bendaharaID int64, posID, tanggal, kategori, keterangan, keteranganPublik string, nominal, notaMediaID int64) (int64, error) {
	if nominal <= 0 {
		return 0, errors.New("nominal harus lebih dari nol")
	}
	if notaMediaID <= 0 {
		return 0, errors.New("foto nota wajib dilampirkan")
	}
	res, err := s.DB.ExecContext(ctx, `INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, nota_media_id, dibuat_oleh, dibuat_at)
		VALUES (?, ?, 'keluar', ?, ?, ?, ?, ?, ?, ?)`,
		posID, tanggal, nominal, kategori, keterangan, keteranganPublik, notaMediaID, bendaharaID, time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
