package konten

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"sakuragakure/internal/db"
)

// Service menangani konten: album, foto, posting kegiatan.
type Service struct {
	Q  *db.Queries
	DB *sql.DB
}

func New(conn *sql.DB) *Service { return &Service{Q: db.New(conn), DB: conn} }

// BuatAlbum membuat album baru beserta foto-fotonya.
func (s *Service) BuatAlbum(ctx context.Context, dibuatOleh int64, judul, tanggal, cerita string, fotoMediaIDs []int64) (int64, error) {
	if judul == "" {
		return 0, errors.New("judul kegiatan wajib diisi")
	}
	if len(fotoMediaIDs) == 0 {
		return 0, errors.New("pilih minimal satu foto")
	}
	slug := Slugify(judul)
	var n int
	_ = s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM album WHERE slug = ?`, slug).Scan(&n)
	if n > 0 {
		// ponytail: suffix waktu cukup untuk skala kecil; ganti UUID bila sering bentrok
		slug = fmt.Sprintf("%s-%d", slug, time.Now().UnixNano()%100000)
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `INSERT INTO album (judul, slug, tanggal, cerita, sampul_media_id, dibuat_oleh) VALUES (?, ?, ?, ?, ?, ?)`,
		judul, slug, tanggal, cerita, fotoMediaIDs[0], dibuatOleh)
	if err != nil {
		return 0, err
	}
	albumID, _ := res.LastInsertId()
	for i, mid := range fotoMediaIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO album_foto (album_id, media_id, keterangan, urutan) VALUES (?, ?, '', ?)`, albumID, mid, i); err != nil {
			return 0, err
		}
	}
	return albumID, tx.Commit()
}

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify membuat slug dari judul.
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "album"
	}
	return s
}
