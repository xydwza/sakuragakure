package backup

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"sakuragakure/internal/db"
)

func TestJalankan(t *testing.T) {
	dir := t.TempDir()
	conn, err := db.Open(filepath.Join(dir, "app.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		t.Fatal(err)
	}

	target, err := Jalankan(conn, dir, 14)
	if err != nil {
		t.Fatalf("jalankan: %v", err)
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("file backup tidak ada: %v", err)
	}
}

func TestPrune(t *testing.T) {
	dir := t.TempDir()
	lama := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	baru := time.Now().Format("2006-01-02")
	for _, n := range []string{"app-" + lama + ".db", "app-" + baru + ".db", "lain.txt"} {
		os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644)
	}
	if err := Prune(dir, 14); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "app-"+lama+".db")); !os.IsNotExist(err) {
		t.Fatalf("cadangan lama harusnya dihapus")
	}
	if _, err := os.Stat(filepath.Join(dir, "app-"+baru+".db")); err != nil {
		t.Fatalf("cadangan baru harusnya tetap ada")
	}
	if _, err := os.Stat(filepath.Join(dir, "lain.txt")); err != nil {
		t.Fatalf("file lain tidak boleh disentuh")
	}
}
