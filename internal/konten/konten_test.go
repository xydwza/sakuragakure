package konten

import (
	"context"
	"path/filepath"
	"testing"

	"sakuragakure/internal/db"
)

func TestBuatAlbum(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`INSERT INTO user (nama, no_wa, created_at) VALUES ('S','+628','2026-10-01T00:00:00+07:00')`); err != nil {
		t.Fatal(err)
	}
	res, _ := conn.Exec(`INSERT INTO media (path, thumb_path, mime, ukuran, akses, dibuat_at) VALUES ('/tmp/a.jpg','/tmp/a_t.jpg','image/jpeg',10,'publik','2026-10-01T00:00:00+07:00')`)
	m1, _ := res.LastInsertId()
	res, _ = conn.Exec(`INSERT INTO media (path, thumb_path, mime, ukuran, akses, dibuat_at) VALUES ('/tmp/b.jpg','/tmp/b_t.jpg','image/jpeg',10,'publik','2026-10-01T00:00:00+07:00')`)
	m2, _ := res.LastInsertId()

	s := New(conn)
	ctx := context.Background()

	// tanpa foto -> ditolak
	if _, err := s.BuatAlbum(ctx, 1, "Kerja bakti", "2026-10-01", "", nil); err == nil {
		t.Fatalf("album tanpa foto harusnya ditolak")
	}

	id, err := s.BuatAlbum(ctx, 1, "Kerja bakti gang 1", "2026-10-01", "cerita", []int64{m1, m2})
	if err != nil {
		t.Fatalf("buat album: %v", err)
	}
	var n int
	conn.QueryRow(`SELECT COUNT(*) FROM album_foto WHERE album_id = ?`, id).Scan(&n)
	if n != 2 {
		t.Fatalf("ingin 2 foto, dapat %d", n)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Kerja Bakti Gang 1!": "kerja-bakti-gang-1",
		"17 Agustus 2026":     "17-agustus-2026",
		"  Spasi  ":           "spasi",
		"!!!":                 "album",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Fatalf("Slugify(%q) = %q, ingin %q", in, got, want)
		}
	}
}
