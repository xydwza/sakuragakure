package rumah

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"sakuragakure/internal/db"
)

const csvUji = `alamat,blok,nomor,gang,status_rumah,nama_kk,gender,umur,jumlah_anggota,bebas_iuran,catatan_validasi
G10/01,G10,01,1,tetap,Nopyan Ari Wibowo,L,42,4,,
G10/02,G10,02,1,tetap,Ipan Sopwan Aliyudin,L,36,1,koordinator gang 1,
G10/03,G10,03,1,kontrak,Alwis,L,,3,,data kk belum lengkap (gender/umur)
G10/14,G10,14,2,kontrak,,,,0,,nama pengontrak belum diisi
G10/11,G10,11,1,tetap,Laura Berton Saragih,L,43,1,,
G10/12,G10,12,1,tetap,Laura Berton Saragih,,,0,rumah kedua (1 rumah dihitung),kk sama dengan G10/11
G21/02,G21,02,1,kosong,,,,0,,
`

func TestBacaCSV(t *testing.T) {
	baris, err := BacaCSV(strings.NewReader(csvUji))
	if err != nil {
		t.Fatalf("baca: %v", err)
	}
	if len(baris) != 7 {
		t.Fatalf("ingin 7 baris, dapat %d", len(baris))
	}
	b := baris[0]
	if b.Alamat != "G10/01" || b.NamaKK != "Nopyan Ari Wibowo" || !b.UmurAda || b.Umur != 42 {
		t.Fatalf("baris pertama salah: %+v", b)
	}
	if baris[2].UmurAda {
		t.Fatalf("G10/03 umur kosong harusnya UmurAda=false")
	}
}

func TestLaporkan(t *testing.T) {
	baris, _ := BacaCSV(strings.NewReader(csvUji))
	l := Laporkan(baris)
	if got := l.DuplikatKK["Laura Berton Saragih"]; len(got) != 2 {
		t.Fatalf("duplikat KK Laura ingin 2 alamat, dapat %v", got)
	}
	if len(l.Catatan) != 3 {
		t.Fatalf("ingin 3 baris catatan validasi, dapat %d", len(l.Catatan))
	}
	if g := l.PerGang[1]; g.Tetap != 4 || g.Kontrak != 1 || g.Kosong != 1 {
		t.Fatalf("ringkasan gang 1 salah: %+v", g)
	}
}

func TestImpor(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer conn.Close()
	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	baris, _ := BacaCSV(strings.NewReader(csvUji))
	h, err := Impor(conn, baris)
	if err != nil {
		t.Fatalf("impor: %v", err)
	}
	if h.Rumah != 7 || h.Penghuni != 6 {
		t.Fatalf("ingin 7 rumah + 6 penghuni, dapat %+v", h)
	}

	// G10/01: umur 42 -> tahun_lahir 1984, perkiraan 1
	var tl, perk int
	var g interface{}
	if err := conn.QueryRow(`SELECT tahun_lahir, tahun_lahir_perkiraan, gender FROM penghuni p JOIN rumah r ON r.id=p.rumah_id WHERE r.alamat='G10/01'`).
		Scan(&tl, &perk, &g); err != nil {
		t.Fatalf("G10/01: %v", err)
	}
	if tl != 1984 || perk != 1 || g != "L" {
		t.Fatalf("G10/01: tl=%d perk=%d gender=%v", tl, perk, g)
	}

	// G10/03: umur kosong -> tahun_lahir NULL, perkiraan 0 (gender tetap L)
	var tlN sql.NullInt64
	if err := conn.QueryRow(`SELECT gender, tahun_lahir, tahun_lahir_perkiraan FROM penghuni p JOIN rumah r ON r.id=p.rumah_id WHERE r.alamat='G10/03'`).
		Scan(&g, &tlN, &perk); err != nil {
		t.Fatalf("G10/03: %v", err)
	}
	if g != "L" || tlN.Valid || perk != 0 {
		t.Fatalf("G10/03: gender=%v tl=%v perk=%d", g, tlN, perk)
	}

	// G10/12: gender kosong -> NULL, bebas iuran tersimpan
	if err := conn.QueryRow(`SELECT gender FROM penghuni p JOIN rumah r ON r.id=p.rumah_id WHERE r.alamat='G10/12'`).Scan(&g); err != nil {
		t.Fatalf("G10/12: %v", err)
	}
	if g != nil {
		t.Fatalf("G10/12: gender harusnya NULL, dapat %v", g)
	}
	var bebas string
	if err := conn.QueryRow(`SELECT bebas_iuran_alasan FROM rumah WHERE alamat='G10/12'`).Scan(&bebas); err != nil {
		t.Fatalf("G10/12 bebas: %v", err)
	}
	if bebas != "rumah kedua (1 rumah dihitung)" {
		t.Fatalf("G10/12 bebas_iuran_alasan=%q", bebas)
	}

	// G10/14: kontrak tanpa nama -> "Penghuni G10/14"
	var nama string
	if err := conn.QueryRow(`SELECT nama_kk FROM penghuni p JOIN rumah r ON r.id=p.rumah_id WHERE r.alamat='G10/14'`).Scan(&nama); err != nil {
		t.Fatalf("G10/14: %v", err)
	}
	if nama != "Penghuni G10/14" {
		t.Fatalf("G10/14: nama=%q", nama)
	}

	// G21/02 kosong -> tanpa penghuni
	var n int
	if err := conn.QueryRow(`SELECT count(*) FROM penghuni p JOIN rumah r ON r.id=p.rumah_id WHERE r.alamat='G21/02'`).Scan(&n); err != nil {
		t.Fatalf("G21/02: %v", err)
	}
	if n != 0 {
		t.Fatalf("G21/02 kosong harusnya tanpa penghuni, dapat %d", n)
	}

	// idempoten: impor ulang tidak duplikasi
	h2, err := Impor(conn, baris)
	if err != nil {
		t.Fatalf("impor ulang: %v", err)
	}
	if h2.Rumah != 0 || h2.Penghuni != 0 || h2.Dilewati != 7 {
		t.Fatalf("impor ulang harusnya 0 baru, dapat %+v", h2)
	}
}
