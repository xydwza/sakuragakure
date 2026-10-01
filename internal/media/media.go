package media

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"

	"sakuragakure/internal/db"
)

const fullMax = 1920
const thumbMax = 480
const jpegQuality = 82

// Service menangani upload gambar dan penyimpanan media.
type Service struct {
	Q   *db.Queries
	DB  *sql.DB
	Dir string
}

func New(conn *sql.DB, dir string) *Service {
	return &Service{Q: db.New(conn), DB: conn, Dir: dir}
}

// Upload menyimpan gambar, re-encode JPEG (membuang EXIF/GPS), membuat
// thumbnail, dan mencatat baris media. Mengembalikan id media.
func (s *Service) Upload(ctx context.Context, data []byte, akses string, pemilik int64) (int64, error) {
	img, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, fmt.Errorf("gambar tidak terbaca: %w", err)
	}

	full := img
	if full.Bounds().Dx() > fullMax || full.Bounds().Dy() > fullMax {
		full = imaging.Fit(img, fullMax, fullMax, imaging.Lanczos)
	}
	var fullBuf bytes.Buffer
	if err := imaging.Encode(&fullBuf, full, imaging.JPEG, imaging.JPEGQuality(jpegQuality)); err != nil {
		return 0, err
	}

	thumb := imaging.Fit(img, thumbMax, thumbMax, imaging.Lanczos)
	var thumbBuf bytes.Buffer
	if err := imaging.Encode(&thumbBuf, thumb, imaging.JPEG, imaging.JPEGQuality(jpegQuality)); err != nil {
		return 0, err
	}

	if err := os.MkdirAll(s.Dir, 0o755); err != nil {
		return 0, err
	}
	name := randHex(16)
	fullPath := filepath.Join(s.Dir, name+".jpg")
	thumbPath := filepath.Join(s.Dir, name+"_thumb.jpg")
	if err := os.WriteFile(fullPath, fullBuf.Bytes(), 0o644); err != nil {
		return 0, err
	}
	if err := os.WriteFile(thumbPath, thumbBuf.Bytes(), 0o644); err != nil {
		return 0, err
	}

	var pemilikVal interface{}
	if pemilik != 0 {
		pemilikVal = pemilik
	}
	res, err := s.DB.ExecContext(ctx, `INSERT INTO media (path, thumb_path, mime, ukuran, akses, pemilik_user_id, dibuat_at)
		VALUES (?, ?, 'image/jpeg', ?, ?, ?, ?)`,
		fullPath, thumbPath, fullBuf.Len(), akses, pemilikVal, time.Now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Get mengambil baris media.
func (s *Service) Get(ctx context.Context, id int64) (db.Medium, error) {
	return s.Q.MediaByID(ctx, id)
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
