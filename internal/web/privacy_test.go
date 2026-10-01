package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"sakuragakure/internal/auth"
	"sakuragakure/internal/db"
)

// TestPrivasiPublik memastikan halaman publik tidak memuat nama KK penghuni,
// termasuk yang tersimpan di keterangan (internal) mutasi.
func TestPrivasiPublik(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		t.Fatal(err)
	}

	if _, err := conn.Exec(`INSERT INTO user (nama, no_wa, created_at) VALUES ('Bendahara','+628','2026-10-01T00:00:00+07:00')`); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`INSERT INTO rumah (alamat, blok, nomor, gang, status) VALUES ('G10/01','G10','01',1,'tetap')`); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`INSERT INTO penghuni (rumah_id, nama_kk, status_huni, mulai) VALUES (1, 'KELUARGA_TEST_RAHASIA', 'pemilik', '2026-04-01')`); err != nil {
		t.Fatal(err)
	}
	// mutasi: keterangan (internal) berisi nama, keterangan_publik bersih
	if _, err := conn.Exec(`INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, dibuat_oleh, dibuat_at)
		VALUES ('kas_rt','2026-10-05','masuk',25000,'Iuran','Iuran dari KELUARGA_TEST_RAHASIA','Setoran iuran warga',1,'2026-10-01T00:00:00+07:00')`); err != nil {
		t.Fatal(err)
	}

	a := auth.New(conn, []byte("test-key-123456789012345678901234567890"))
	h := NewHandlers(conn, a, t.TempDir())

	req := func(path string) *http.Request {
		ctx, _ := a.Sessions.Load(context.Background(), "")
		return httptest.NewRequest(http.MethodGet, path, nil).WithContext(ctx)
	}

	cases := map[string]http.HandlerFunc{
		"/":         h.Beranda,
		"/kas":      h.Kas,
		"/kegiatan": h.Kegiatan,
		"/aturan":   h.Aturan,
	}
	for path, handler := range cases {
		rr := httptest.NewRecorder()
		handler(rr, req(path))
		if rr.Code != http.StatusOK {
			t.Fatalf("%s: kode %d", path, rr.Code)
		}
		if strings.Contains(rr.Body.String(), "KELUARGA_TEST_RAHASIA") {
			t.Fatalf("%s bocor nama KK penghuni", path)
		}
	}

	// kontrol positif: kas tetap menampilkan keterangan publik
	rr := httptest.NewRecorder()
	h.Kas(rr, req("/kas"))
	if !strings.Contains(rr.Body.String(), "Setoran iuran warga") {
		t.Fatalf("/kas tidak menampilkan keterangan_publik (uji positif gagal)")
	}
}
