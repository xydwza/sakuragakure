package kas

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"sakuragakure/internal/db"
)

func newKas(t *testing.T) (*Service, *sql.DB) {
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

func nullable(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func TestKelopakPerGang(t *testing.T) {
	s, conn := newKas(t)
	ctx := context.Background()

	if _, err := conn.Exec(`INSERT INTO user (nama, no_wa, created_at) VALUES ('T', '+628', '2026-10-01T00:00:00+07:00')`); err != nil {
		t.Fatal(err)
	}

	tambahRumah := func(alamat string, gang int, status, bebas string) int64 {
		res, err := conn.Exec(`INSERT INTO rumah (alamat, blok, nomor, gang, status, bebas_iuran_alasan) VALUES (?, ?, ?, ?, ?, ?)`,
			alamat, "G10", "01", gang, status, nullable(bebas))
		if err != nil {
			t.Fatalf("rumah: %v", err)
		}
		id, _ := res.LastInsertId()
		return id
	}

	// gang 1: 3 wajib, 1 kosong, 1 bebas
	r1 := tambahRumah("G10/01", 1, "tetap", "")
	r2 := tambahRumah("G10/02", 1, "tetap", "")
	r3 := tambahRumah("G10/03", 1, "kontrak", "")
	tambahRumah("G21/02", 1, "kosong", "")
	tambahRumah("G10/04", 1, "tetap", "bebas")
	// gang 2: 2 wajib
	tambahRumah("G11/01", 2, "tetap", "")
	tambahRumah("G11/02", 2, "kontrak", "")

	kasihan := func(rumahID int64, status string) {
		tr, _ := conn.Exec(`INSERT INTO tagihan (rumah_id, jenis, periode, nominal) VALUES (?, 'kas', '2026-10', 25000)`, rumahID)
		tagihanID, _ := tr.LastInsertId()
		if _, err := conn.Exec(`INSERT INTO pembayaran (tagihan_id, metode, status, dicatat_oleh, dicatat_at) VALUES (?, 'tunai', ?, 1, '2026-10-01T00:00:00+07:00')`, tagihanID, status); err != nil {
			t.Fatalf("pembayaran: %v", err)
		}
	}
	kasihan(r1, "diterima")
	kasihan(r2, "dipegang")
	kasihan(r3, "disetor")

	k, err := s.KelopakPerGang(ctx, "2026-10")
	if err != nil {
		t.Fatalf("kelopak: %v", err)
	}
	if len(k) != 2 {
		t.Fatalf("ingin 2 gang, dapat %d", len(k))
	}
	if k[0].Gang != 1 || k[0].Total != 3 || k[0].Diterima != 1 || k[0].Dipegang != 2 {
		t.Fatalf("gang 1 salah: %+v", k[0])
	}
	if k[1].Gang != 2 || k[1].Total != 2 || k[1].Diterima != 0 || k[1].Dipegang != 0 {
		t.Fatalf("gang 2 salah: %+v", k[1])
	}
}

func TestSaldoPosis(t *testing.T) {
	s, conn := newKas(t)
	ctx := context.Background()
	if _, err := conn.Exec(`INSERT INTO user (nama, no_wa, created_at) VALUES ('T', '+628', '2026-10-01T00:00:00+07:00')`); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, dibuat_oleh, dibuat_at) VALUES ('kas_rt','2026-10-01','masuk',50000,'Iuran','i','i',1,'2026-10-01T00:00:00+07:00')`,
		`INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, ref_tipe, dibuat_oleh, dibuat_at) VALUES ('kas_rt','2026-10-02','keluar',10000,'Alokasi','b','b','alokasi',1,'2026-10-01T00:00:00+07:00')`,
		`INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, dibuat_oleh, dibuat_at) VALUES ('dana_sosial','2026-10-03','masuk',20000,'Alokasi','a','a',1,'2026-10-01T00:00:00+07:00')`,
	} {
		if _, err := conn.Exec(q); err != nil {
			t.Fatalf("mutasi: %v", err)
		}
	}
	m, err := s.SaldoPosis(ctx)
	if err != nil {
		t.Fatalf("saldo: %v", err)
	}
	if m["kas_rt"] != 40000 || m["dana_sosial"] != 20000 {
		t.Fatalf("saldo salah: %+v", m)
	}
}
