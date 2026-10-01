package laporan

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"sakuragakure/internal/db"
)

func newLaporan(t *testing.T) *Service {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := db.Migrate(conn); err != nil {
		t.Fatal(err)
	}
	return New(conn)
}

func TestEkspor(t *testing.T) {
	s := newLaporan(t)
	data, err := s.Data(context.Background())
	if err != nil {
		t.Fatalf("data: %v", err)
	}
	// migrasi historis sudah ada, jadi data > 0
	if len(data) == 0 {
		t.Fatalf("harusnya ada data historis")
	}

	csv, err := CSV(data)
	if err != nil || !bytes.Contains(csv, []byte("Tanggal,Pos,Kategori")) {
		t.Fatalf("csv: %v", err)
	}
	xlsx, err := XLSX(data)
	if err != nil || !bytes.HasPrefix(xlsx, []byte("PK")) { // zip magic
		t.Fatalf("xlsx: %v", err)
	}
	pdf, err := PDF(data)
	if err != nil || !bytes.Contains(pdf, []byte("%PDF")) {
		t.Fatalf("pdf: %v", err)
	}
	if !strings.Contains(string(csv), "TOTAL") {
		t.Fatalf("csv tanpa baris total")
	}
}
