package backup

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Jalankan membuat cadangan VACUUM INTO dan membuang yang lebih tua dari simpanHari.
func Jalankan(db *sql.DB, dir string, simpanHari int) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	nama := "app-" + time.Now().Format("2006-01-02") + ".db"
	target := filepath.Join(dir, nama)
	if _, err := db.Exec(fmt.Sprintf(`VACUUM INTO '%s'`, strings.ReplaceAll(target, "'", ""))); err != nil {
		return "", err
	}
	if err := Prune(dir, simpanHari); err != nil {
		return target, err
	}
	return target, nil
}

// Prune menghapus cadangan app-*.db yang lebih tua dari simpanHari.
func Prune(dir string, simpanHari int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	batas := time.Now().AddDate(0, 0, -simpanHari)
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "app-") || !strings.HasSuffix(name, ".db") {
			continue
		}
		t, err := time.Parse("2006-01-02", strings.TrimSuffix(strings.TrimPrefix(name, "app-"), ".db"))
		if err != nil {
			continue
		}
		if t.Before(batas) {
			_ = os.Remove(filepath.Join(dir, name))
		}
	}
	return nil
}
