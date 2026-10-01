package auth

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
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

func TestBuatUserPassword(t *testing.T) {
	conn := newTestDB(t)
	id, err := BuatUser(conn, "Bendahara", "+628123", "bendahara", "rahasia123", 0)
	if err != nil {
		t.Fatalf("buat user: %v", err)
	}
	var hash string
	if err := conn.QueryRow(`SELECT password_hash FROM user WHERE id = ?`, id).Scan(&hash); err != nil {
		t.Fatal(err)
	}
	if !cekPassword(hash, "rahasia123") {
		t.Fatalf("kata sandi benar harusnya cocok")
	}
	if cekPassword(hash, "salah") {
		t.Fatalf("kata sandi salah harusnya gagal")
	}
	// no_wa dinormalisasi
	var noWA string
	conn.QueryRow(`SELECT no_wa FROM user WHERE id = ?`, id).Scan(&noWA)
	if noWA != "+628123" {
		t.Fatalf("no_wa ingin +628123, dapat %q", noWA)
	}
}

func TestLogin(t *testing.T) {
	conn := newTestDB(t)
	a := New(conn, []byte("test-key-123456789012345678901234567890"))
	if _, err := BuatUser(conn, "Admin", "+628123", "admin", "pass123", 0); err != nil {
		t.Fatal(err)
	}

	post := func(noWA, pw string) *httptest.ResponseRecorder {
		form := url.Values{"no_wa": {noWA}, "password": {pw}}
		req := httptest.NewRequest(http.MethodPost, "/masuk", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		a.Login(rr, req)
		return rr
	}

	// login benar -> 303 + cookie token
	rr := post("+628123", "pass123")
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("login benar ingin 303, dapat %d", rr.Code)
	}
	found := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == tokenCookie && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Fatalf("login benar harusnya menerbitkan cookie token")
	}

	// login salah -> 401
	if rr := post("+628123", "salah"); rr.Code != http.StatusUnauthorized {
		t.Fatalf("login salah ingin 401, dapat %d", rr.Code)
	}
}
