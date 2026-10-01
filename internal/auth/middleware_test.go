package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func ctxUser(uid int64) context.Context {
	if uid == 0 {
		return context.Background()
	}
	return context.WithValue(context.Background(), ctxKey{}, uid)
}

func TestRequirePeran(t *testing.T) {
	conn := newTestDB(t)
	a := New(conn, []byte("test-key-123456789012345678901234567890"))

	wargaID := seedUser(t, conn, "Warga", "+628100", "warga")
	bendaharaID := seedUser(t, conn, "Bendahara", "+628200", "bendahara")

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := a.RequirePeran("bendahara")(ok)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctxUser(0)))
	if rr.Code != http.StatusSeeOther || rr.Header().Get("Location") != "/masuk" {
		t.Fatalf("tamu: ingin 303 ke /masuk, dapat %d %q", rr.Code, rr.Header().Get("Location"))
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctxUser(wargaID)))
	if rr.Code != http.StatusForbidden {
		t.Fatalf("peran warga: ingin 403, dapat %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctxUser(bendaharaID)))
	if rr.Code != http.StatusOK {
		t.Fatalf("peran bendahara: ingin 200, dapat %d", rr.Code)
	}
}

func TestRequireLogin(t *testing.T) {
	conn := newTestDB(t)
	a := New(conn, []byte("test-key-123456789012345678901234567890"))
	uid := seedUser(t, conn, "Warga", "+628100", "warga")

	h := a.RequireLogin(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctxUser(0)))
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("tamu: ingin 303, dapat %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctxUser(uid)))
	if rr.Code != http.StatusOK {
		t.Fatalf("masuk: ingin 200, dapat %d", rr.Code)
	}
}

func TestPunyaGang(t *testing.T) {
	conn := newTestDB(t)
	a := New(conn, []byte("test-key-123456789012345678901234567890"))

	res, _ := conn.Exec(`INSERT INTO user (nama, no_wa, created_at) VALUES ('K1','+628100','2026-10-01T00:00:00+07:00')`)
	uid, _ := res.LastInsertId()
	if _, err := conn.Exec(`INSERT INTO user_peran (user_id, peran, gang) VALUES (?, 'koordinator', 1)`, uid); err != nil {
		t.Fatal(err)
	}

	ctx := ctxUser(uid)
	if !a.PunyaGang(ctx, 1, "koordinator", "pembantu_koordinator") {
		t.Fatalf("koordinator gang 1 harus punya akses gang 1")
	}
	if a.PunyaGang(ctx, 2, "koordinator", "pembantu_koordinator") {
		t.Fatalf("koordinator gang 1 tidak boleh punya akses gang 2")
	}
}

func TestJWTRoundTrip(t *testing.T) {
	conn := newTestDB(t)
	a := New(conn, []byte("test-key-123456789012345678901234567890"))
	_ = conn

	token, err := a.TerbitkanToken(42)
	if err != nil {
		t.Fatalf("terbitkan: %v", err)
	}
	uid, err := a.verifikasiToken(token)
	if err != nil || uid != 42 {
		t.Fatalf("verifikasi: uid=%d err=%v", uid, err)
	}
	if _, err := a.verifikasiToken(token + "tamper"); err == nil {
		t.Fatalf("token dirusak harusnya gagal")
	}
}
