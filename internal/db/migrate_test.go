package db

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateAndTriggers(t *testing.T) {
	conn, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	if err := Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// pastikan seed utuh (multi-baris INSERT tidak terpotong parser goose)
	for _, c := range []struct{ q, want string }{
		{`SELECT count(*) FROM pos_dana`, "6"},
		{`SELECT count(*) FROM setting`, "18"},
		{`SELECT count(*) FROM aturan_pasal`, "11"},
		{`SELECT count(*) FROM inventaris`, "4"},
		{`SELECT count(*) FROM jadwal_rutin`, "6"},
	} {
		var got string
		if err := conn.QueryRow(c.q).Scan(&got); err != nil {
			t.Fatalf("%s: %v", c.q, err)
		}
		if got != c.want {
			t.Fatalf("%s: dapat %s, ingin %s", c.q, got, c.want)
		}
	}

	// user dibutuhkan untuk FK mutasi.dibuat_oleh dan periode_tutup.ditutup_oleh
	if _, err := conn.Exec(`INSERT INTO user (nama, no_wa, created_at) VALUES ('Tester','6281111','2026-10-01T00:00:00+07:00')`); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	// 1. periode tertutup mengunci insert mutasi
	if _, err := conn.Exec(`INSERT INTO periode_tutup (periode, ditutup_oleh, ditutup_at, snapshot_saldo) VALUES ('2026-04', 1, '2026-05-01T00:00:00+07:00', '{}')`); err != nil {
		t.Fatalf("insert periode_tutup: %v", err)
	}
	if _, err := conn.Exec(mutasiInsert("kas_rt", "2026-04-10")); err == nil || !strings.Contains(err.Error(), "periode sudah ditutup") {
		t.Fatalf("insert mutasi periode tertutup: ingin error kunci, dapat %v", err)
	}

	// 2. mutasi di periode terbuka bisa masuk, tapi tidak bisa di-UPDATE/DELETE
	res, err := conn.Exec(mutasiInsert("kas_rt", "2026-10-15"))
	if err != nil {
		t.Fatalf("insert mutasi terbuka: %v", err)
	}
	id, _ := res.LastInsertId()

	if _, err := conn.Exec(`UPDATE mutasi SET nominal = 999 WHERE id = ?`, id); err == nil || !strings.Contains(err.Error(), "mutasi tidak boleh diubah") {
		t.Fatalf("update mutasi: ingin error, dapat %v", err)
	}
	if _, err := conn.Exec(`DELETE FROM mutasi WHERE id = ?`, id); err == nil || !strings.Contains(err.Error(), "mutasi tidak boleh dihapus") {
		t.Fatalf("delete mutasi: ingin error, dapat %v", err)
	}

	// 3. periode_tutup tidak bisa dibuka (dihapus)
	if _, err := conn.Exec(`DELETE FROM periode_tutup WHERE periode = '2026-04'`); err == nil || !strings.Contains(err.Error(), "periode tertutup tidak bisa dibuka") {
		t.Fatalf("delete periode_tutup: ingin error, dapat %v", err)
	}

	// 4. pengeluaran (arah keluar) tanpa nota dan bukan alokasi/koreksi/impor gagal di CHECK
	if _, err := conn.Exec(`INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, dibuat_oleh, dibuat_at) VALUES ('kas_rt','2026-10-16','keluar',50000,'ATK','beli','beli',1,'2026-10-01T00:00:00+07:00')`); err == nil || !strings.Contains(err.Error(), "CHECK constraint failed") {
		t.Fatalf("keluar tanpa nota: ingin CHECK gagal, dapat %v", err)
	}
}

func mutasiInsert(pos, tanggal string) string {
	return `INSERT INTO mutasi (pos_id, tanggal, arah, nominal, kategori, keterangan, keterangan_publik, dibuat_oleh, dibuat_at) VALUES ('` +
		pos + `','` + tanggal + `','masuk',25000,'Iuran','iuran','iuran',1,'2026-10-01T00:00:00+07:00')`
}
