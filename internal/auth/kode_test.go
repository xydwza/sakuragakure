package auth

import "testing"

func TestKodeCadangan(t *testing.T) {
	conn := newTestDB(t)
	uid := seedUser(t, conn, "A", "+628123", "warga")
	kode, err := BuatKodeCadangan(conn, uid, uid)
	if err != nil {
		t.Fatalf("buat: %v", err)
	}
	got, err := PakaiKodeCadangan(conn, kode)
	if err != nil {
		t.Fatalf("pakai: %v", err)
	}
	if got != uid {
		t.Fatalf("ingin user %d, dapat %d", uid, got)
	}
	if _, err := PakaiKodeCadangan(conn, kode); err != ErrKodeSalah {
		t.Fatalf("kode sudah dipakai harusnya ErrKodeSalah, dapat %v", err)
	}
}

func TestKodeCadanganSalah(t *testing.T) {
	conn := newTestDB(t)
	if _, err := PakaiKodeCadangan(conn, "XXXXXXXX"); err != ErrKodeSalah {
		t.Fatalf("kode salah harusnya ErrKodeSalah, dapat %v", err)
	}
}

func TestKodeCadanganKedaluwarsa(t *testing.T) {
	conn := newTestDB(t)
	uid := seedUser(t, conn, "A", "+628123", "warga")
	kode, _ := BuatKodeCadangan(conn, uid, uid)
	if _, err := conn.Exec(`UPDATE kode_cadangan SET kedaluwarsa = '2020-01-01T00:00:00+07:00'`); err != nil {
		t.Fatal(err)
	}
	if _, err := PakaiKodeCadangan(conn, kode); err != ErrKodeKedaluwarsa {
		t.Fatalf("ingin ErrKodeKedaluwarsa, dapat %v", err)
	}
}
