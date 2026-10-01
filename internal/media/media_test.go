package media

import (
	"bytes"
	"context"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/disintegration/imaging"

	"sakuragakure/internal/db"
)

func newMedia(t *testing.T) *Service {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return New(conn, filepath.Join(t.TempDir(), "media"))
}

func TestUpload(t *testing.T) {
	s := newMedia(t)
	ctx := context.Background()

	// gambar uji 1000x800 PNG
	img := imaging.New(1000, 800, color.NRGBA{R: 200, G: 50, B: 80, A: 255})
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.PNG); err != nil {
		t.Fatal(err)
	}

	id, err := s.Upload(ctx, buf.Bytes(), "warga", 0)
	if err != nil {
		t.Fatalf("upload: %v", err)
	}
	m, err := s.Get(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if m.Mime != "image/jpeg" {
		t.Fatalf("mime ingin image/jpeg, dapat %s", m.Mime)
	}
	if m.Akses != "warga" {
		t.Fatalf("akses ingin warga, dapat %s", m.Akses)
	}

	full, err := os.Stat(m.Path)
	if err != nil {
		t.Fatalf("file full tidak ada: %v", err)
	}
	thumb, err := os.Stat(m.ThumbPath.String)
	if err != nil {
		t.Fatalf("file thumb tidak ada: %v", err)
	}
	if thumb.Size() >= full.Size() {
		t.Fatalf("thumb harus lebih kecil dari full")
	}

	// verifikasi thumb benar-benar diperkecil
	f, _ := imaging.Open(m.Path)
	if f.Bounds().Dx() != 1000 {
		t.Fatalf("full ingin 1000px, dapat %d", f.Bounds().Dx())
	}
	tm, _ := imaging.Open(m.ThumbPath.String)
	if tm.Bounds().Dx() != 480 {
		t.Fatalf("thumb ingin 480px, dapat %d", tm.Bounds().Dx())
	}
}

// TestUploadGambarBesar mengecek downscale full ke 1920.
func TestUploadGambarBesar(t *testing.T) {
	s := newMedia(t)
	ctx := context.Background()
	img := imaging.New(4000, 3000, color.NRGBA{R: 10, G: 10, B: 10, A: 255})
	var buf bytes.Buffer
	_ = imaging.Encode(&buf, img, imaging.PNG)

	id, err := s.Upload(ctx, buf.Bytes(), "pengurus", 0)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := s.Get(ctx, id)
	f, _ := imaging.Open(m.Path)
	if f.Bounds().Dx() > 1920 {
		t.Fatalf("full ingin <=1920, dapat %d", f.Bounds().Dx())
	}
}
