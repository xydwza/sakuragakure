package auth

import (
	"database/sql"
	"path/filepath"
	"testing"

	"sakuragakure/internal/db"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return conn
}

func seedUser(t *testing.T, conn *sql.DB, nama, noWA, peran string) int64 {
	t.Helper()
	res, err := conn.Exec(`INSERT INTO user (nama, no_wa, aktif, created_at) VALUES (?, ?, 1, '2026-10-01T00:00:00+07:00')`, nama, noWA)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	id, _ := res.LastInsertId()
	if _, err := conn.Exec(`INSERT INTO user_peran (user_id, peran, gang, urutan) VALUES (?, ?, 0, 0)`, id, peran); err != nil {
		t.Fatalf("seed peran: %v", err)
	}
	return id
}
