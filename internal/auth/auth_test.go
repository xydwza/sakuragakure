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
	id, err := BuatUser(conn, "Bendahara", "bendahara", "+628123", "bendahara", "", 0)
	if err != nil {
		t.Fatalf("buat user: %v", err)
	}
	var hash, username string
	if err := conn.QueryRow(`SELECT password_hash, username FROM user WHERE id = ?`, id).Scan(&hash, &username); err != nil {
		t.Fatal(err)
	}
	if username != "bendahara" {
		t.Fatalf("username ingin bendahara, dapat %q", username)
	}
	// password default = username
	if !cekPassword(hash, "bendahara") {
		t.Fatalf("kata sandi default harusnya = username")
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
	if _, err := BuatUser(conn, "Admin", "admin", "+628123", "admin", "", 0); err != nil {
		t.Fatal(err)
	}

	post := func(user, pw string) *httptest.ResponseRecorder {
		form := url.Values{"user": {user}, "password": {pw}}
		req := httptest.NewRequest(http.MethodPost, "/masuk", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		a.Login(rr, req)
		return rr
	}

	// login via username -> 303 + cookie token
	rr := post("admin", "admin")
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("login username ingin 303, dapat %d", rr.Code)
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

	// login via nomor WA juga bisa
	if rr := post("+628123", "admin"); rr.Code != http.StatusSeeOther {
		t.Fatalf("login via WA ingin 303, dapat %d", rr.Code)
	}

	// login salah -> 401
	if rr := post("admin", "salah"); rr.Code != http.StatusUnauthorized {
		t.Fatalf("login salah ingin 401, dapat %d", rr.Code)
	}
}
