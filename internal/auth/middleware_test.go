package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func reqSesi(a *Auth, t *testing.T, uid int64) *http.Request {
	t.Helper()
	ctx, err := a.Sessions.Load(context.Background(), "")
	if err != nil {
		t.Fatalf("load sesi: %v", err)
	}
	if uid != 0 {
		a.Sessions.Put(ctx, "userID", uid)
	}
	return httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
}

func TestRequirePeran(t *testing.T) {
	conn := newTestDB(t)
	a := New(conn, []byte("test-key-123456789012345678901234567890"))

	wargaID := seedUser(t, conn, "Warga", "+628100", "warga")
	bendaharaID := seedUser(t, conn, "Bendahara", "+628200", "bendahara")

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := a.RequirePeran("bendahara")(ok)

	// tamu -> redirect ke /masuk
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, reqSesi(a, t, 0))
	if rr.Code != http.StatusSeeOther || rr.Header().Get("Location") != "/masuk" {
		t.Fatalf("tamu: ingin 303 ke /masuk, dapat %d %q", rr.Code, rr.Header().Get("Location"))
	}

	// peran salah -> 403
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, reqSesi(a, t, wargaID))
	if rr.Code != http.StatusForbidden {
		t.Fatalf("peran warga: ingin 403, dapat %d", rr.Code)
	}

	// peran benar -> 200
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, reqSesi(a, t, bendaharaID))
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
	h.ServeHTTP(rr, reqSesi(a, t, 0))
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("tamu: ingin 303, dapat %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, reqSesi(a, t, uid))
	if rr.Code != http.StatusOK {
		t.Fatalf("masuk: ingin 200, dapat %d", rr.Code)
	}
}

func TestPunyaGang(t *testing.T) {
	conn := newTestDB(t)
	a := New(conn, []byte("test-key-123456789012345678901234567890"))

	// koordinator gang 1
	res, _ := conn.Exec(`INSERT INTO user (nama, no_wa, created_at) VALUES ('K1','+628100','2026-10-01T00:00:00+07:00')`)
	uid, _ := res.LastInsertId()
	if _, err := conn.Exec(`INSERT INTO user_peran (user_id, peran, gang) VALUES (?, 'koordinator', 1)`, uid); err != nil {
		t.Fatal(err)
	}

	ctx, _ := a.Sessions.Load(context.Background(), "")
	a.Sessions.Put(ctx, "userID", uid)

	if !a.PunyaGang(ctx, 1, "koordinator", "pembantu_koordinator") {
		t.Fatalf("koordinator gang 1 harus punya akses gang 1")
	}
	if a.PunyaGang(ctx, 2, "koordinator", "pembantu_koordinator") {
		t.Fatalf("koordinator gang 1 tidak boleh punya akses gang 2")
	}
}
