package iuran

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"sakuragakure/internal/db"
)

func newIuran(t *testing.T) (*Service, *sql.DB) {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return New(conn), conn
}

func seedUser(t *testing.T, conn *sql.DB, nama, noWA string) int64 {
	res, err := conn.Exec(`INSERT INTO user (nama, no_wa, created_at) VALUES (?, ?, '2026-10-01T00:00:00+07:00')`, nama, noWA)
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func seedRumah(t *testing.T, conn *sql.DB, alamat string, gang int, status, bebas string) int64 {
	var b interface{}
	if bebas != "" {
		b = bebas
	}
	res, err := conn.Exec(`INSERT INTO rumah (alamat, blok, nomor, gang, status, bebas_iuran_alasan) VALUES (?, 'G10', '01', ?, ?, ?)`, alamat, gang, status, b)
	if err != nil {
		t.Fatalf("rumah: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func TestAlurIuranLengkap(t *testing.T) {
	s, conn := newIuran(t)
	ctx := context.Background()
	koord := seedUser(t, conn, "Koord", "+628100")
	bendahara := seedUser(t, conn, "Bendahara", "+628200")

	r1 := seedRumah(t, conn, "G10/01", 1, "tetap", "")
	r2 := seedRumah(t, conn, "G10/02", 1, "kontrak", "")
	seedRumah(t, conn, "G10/03", 1, "kosong", "")
	seedRumah(t, conn, "G10/04", 1, "tetap", "bebas")

	// tagihan otomatis dibuat oleh toggle; tapi uji GenerateTagihan dulu
	n, err := s.GenerateTagihan(ctx, "2026-10")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if n != 2 {
		t.Fatalf("ingin 2 tagihan (2 wajib), dapat %d", n)
	}

	// tap rumah kosong -> error
	if _, err := s.TogglePembayaran(ctx, koord, "G10/03", "2026-10"); err == nil {
		t.Fatalf("kosong harusnya ditolak")
	}
	// tap rumah bebas -> error
	if _, err := s.TogglePembayaran(ctx, koord, "G10/04", "2026-10"); err == nil {
		t.Fatalf("bebas harusnya ditolak")
	}

	// tap wajib -> dipegang
	st, err := s.TogglePembayaran(ctx, koord, "G10/01", "2026-10")
	if err != nil || st != "dipegang" {
		t.Fatalf("tap G10/01: %v (%s)", err, st)
	}
	// tap lagi -> batal
	st, err = s.TogglePembayaran(ctx, koord, "G10/01", "2026-10")
	if err != nil || st != "batal" {
		t.Fatalf("tap ulang G10/01: %v (%s)", err, st)
	}

	// tap 2 rumah -> dipegang
	for _, a := range []string{"G10/01", "G10/02"} {
		if _, err := s.TogglePembayaran(ctx, koord, a, "2026-10"); err != nil {
			t.Fatalf("tap %s: %v", a, err)
		}
	}

	// setor -> menunggu, pembayaran disetor
	setoranID, total, err := s.Setor(ctx, koord, 1, "2026-10")
	if err != nil {
		t.Fatalf("setor: %v", err)
	}
	if total != 50000 {
		t.Fatalf("total ingin 50000, dapat %d", total)
	}

	// toggle setelah disetor -> error
	if _, err := s.TogglePembayaran(ctx, koord, "G10/01", "2026-10"); err == nil {
		t.Fatalf("toggle setelah disetor harusnya ditolak")
	}

	// terima -> diterima + mutasi
	if err := s.TerimaSetoran(ctx, bendahara, setoranID); err != nil {
		t.Fatalf("terima: %v", err)
	}
	var saldo int64
	conn.QueryRow(`SELECT COALESCE(SUM(CASE WHEN arah='masuk' THEN nominal ELSE -nominal END),0) FROM mutasi WHERE pos_id='kas_rt'`).Scan(&saldo)
	if saldo != 50000 {
		t.Fatalf("saldo kas_rt ingin 50000, dapat %d", saldo)
	}
	// toggle setelah diterima -> error
	if _, err := s.TogglePembayaran(ctx, koord, "G10/01", "2026-10"); err == nil {
		t.Fatalf("toggle setelah diterima harusnya ditolak")
	}
	// terima lagi -> error
	if err := s.TerimaSetoran(ctx, bendahara, setoranID); err == nil {
		t.Fatalf("terima kedua harusnya ditolak")
	}
	_ = r1
	_ = r2
}

func TestTolakSetoran(t *testing.T) {
	s, conn := newIuran(t)
	ctx := context.Background()
	koord := seedUser(t, conn, "Koord", "+628100")
	bendahara := seedUser(t, conn, "Bendahara", "+628200")
	seedRumah(t, conn, "G10/01", 1, "tetap", "")

	if _, err := s.TogglePembayaran(ctx, koord, "G10/01", "2026-10"); err != nil {
		t.Fatal(err)
	}
	setoranID, _, err := s.Setor(ctx, koord, 1, "2026-10")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.TolakSetoran(ctx, bendahara, setoranID, "uang kurang"); err != nil {
		t.Fatal(err)
	}

	// pembayaran kembali dipegang
	var st string
	conn.QueryRow(`SELECT status FROM pembayaran`).Scan(&st)
	if st != "dipegang" {
		t.Fatalf("setelah tolak ingin dipegang, dapat %s", st)
	}
	// tidak ada mutasi masuk
	var n int
	conn.QueryRow(`SELECT COUNT(*) FROM mutasi`).Scan(&n)
	if n != 0 {
		t.Fatalf("tolak tidak boleh membuat mutasi, dapat %d", n)
	}
}
