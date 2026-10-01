package rumah

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

const tahunImpor = 2026
const mulaiDefault = "2026-04-01"

type Baris struct {
	Alamat, Blok, Nomor string
	Gang                int
	Status              string // tetap | kontrak | kosong
	NamaKK, Gender      string
	Umur                int
	UmurAda             bool
	JumlahAnggota       int
	BebasIuran          string
	Catatan             string
}

func (b Baris) WajibIuran() bool { return b.Status != "kosong" && b.BebasIuran == "" }

func (b Baris) NamaPenghuni() string {
	if b.NamaKK != "" {
		return b.NamaKK
	}
	return "Penghuni " + b.Alamat
}

// BacaCSV membaca CSV warga (header: alamat,blok,nomor,gang,status_rumah,nama_kk,...).
func BacaCSV(r io.Reader) ([]Baris, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	rows, err := cr.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return nil, fmt.Errorf("csv kosong")
	}
	idx := func(name string) int {
		for i, h := range rows[0] {
			if h == name {
				return i
			}
		}
		return -1
	}
	iA, iB, iN, iG := idx("alamat"), idx("blok"), idx("nomor"), idx("gang")
	iS, iK, iGe, iU := idx("status_rumah"), idx("nama_kk"), idx("gender"), idx("umur")
	iJ, iBe, iC := idx("jumlah_anggota"), idx("bebas_iuran"), idx("catatan_validasi")
	if iA < 0 || iS < 0 {
		return nil, fmt.Errorf("header csv kurang kolom alamat/status_rumah")
	}
	var out []Baris
	for _, row := range rows[1:] {
		if iA >= len(row) || strings.TrimSpace(row[iA]) == "" {
			continue
		}
		get := func(i int) string {
			if i < 0 || i >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[i])
		}
		b := Baris{
			Alamat: row[iA], Blok: get(iB), Nomor: get(iN),
			Status: get(iS), NamaKK: get(iK), Gender: get(iGe),
			BebasIuran: get(iBe), Catatan: get(iC),
		}
		b.Gang, _ = strconv.Atoi(get(iG))
		b.JumlahAnggota, _ = strconv.Atoi(get(iJ))
		if u := get(iU); u != "" {
			if n, err := strconv.Atoi(u); err == nil {
				b.Umur, b.UmurAda = n, true
			}
		}
		out = append(out, b)
	}
	return out, nil
}

type RingkasanGang struct {
	Tetap, Kontrak, Kosong int
}

type Laporan struct {
	PerGang    map[int]RingkasanGang
	DuplikatKK map[string][]string // nama KK -> daftar alamat (len > 1)
	Catatan    []Baris             // baris dengan catatan_validasi
}

// Laporkan menghitung ringkasan untuk dry-run.
func Laporkan(baris []Baris) Laporan {
	l := Laporan{PerGang: map[int]RingkasanGang{}, DuplikatKK: map[string][]string{}}
	perNama := map[string][]string{}
	for _, b := range baris {
		g := l.PerGang[b.Gang]
		switch b.Status {
		case "tetap":
			g.Tetap++
		case "kontrak":
			g.Kontrak++
		default:
			g.Kosong++
		}
		l.PerGang[b.Gang] = g
		if b.NamaKK != "" {
			perNama[b.NamaKK] = append(perNama[b.NamaKK], b.Alamat)
		}
		if b.Catatan != "" {
			l.Catatan = append(l.Catatan, b)
		}
	}
	for nama, alamat := range perNama {
		if len(alamat) > 1 {
			sort.Strings(alamat)
			l.DuplikatKK[nama] = alamat
		}
	}
	return l
}

func (l Laporan) String() string {
	var sb strings.Builder
	gangs := make([]int, 0, len(l.PerGang))
	for g := range l.PerGang {
		gangs = append(gangs, g)
	}
	sort.Ints(gangs)
	total := 0
	for _, g := range gangs {
		s := l.PerGang[g]
		n := s.Tetap + s.Kontrak + s.Kosong
		total += n
		fmt.Fprintf(&sb, "Gang %d: %d rumah (tetap %d, kontrak %d, kosong %d)\n", g, n, s.Tetap, s.Kontrak, s.Kosong)
	}
	fmt.Fprintf(&sb, "Total: %d rumah\n", total)

	fmt.Fprintf(&sb, "\nDuplikasi KK lintas rumah: %d\n", len(l.DuplikatKK))
	names := make([]string, 0, len(l.DuplikatKK))
	for n := range l.DuplikatKK {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(&sb, "  %s: %s\n", n, strings.Join(l.DuplikatKK[n], ", "))
	}

	fmt.Fprintf(&sb, "\nBaris dengan catatan validasi: %d\n", len(l.Catatan))
	for _, c := range l.Catatan {
		fmt.Fprintf(&sb, "  %s: %s\n", c.Alamat, c.Catatan)
	}
	return sb.String()
}

type HasilImpor struct {
	Rumah, Penghuni int
	Dilewati        int
}

// Impor menulis rumah dan penghuni aktif ke DB. Idempotent: rumah yang sudah
// ada dilewati, penghuni aktif yang sudah ada tidak diduplikasi.
func Impor(db *sql.DB, baris []Baris) (HasilImpor, error) {
	var h HasilImpor
	tx, err := db.Begin()
	if err != nil {
		return h, err
	}
	defer tx.Rollback()

	for _, b := range baris {
		bebas := nullable(b.BebasIuran)
		res, err := tx.Exec(`INSERT INTO rumah (alamat, blok, nomor, gang, status, bebas_iuran_alasan, catatan_validasi)
			VALUES (?,?,?,?,?,?,?) ON CONFLICT(alamat) DO NOTHING`,
			b.Alamat, b.Blok, b.Nomor, b.Gang, b.Status, bebas, nullable(b.Catatan))
		if err != nil {
			return h, err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			h.Dilewati++
		} else {
			h.Rumah++
		}

		if b.Status == "kosong" {
			continue
		}

		var rumahID int64
		if err := tx.QueryRow(`SELECT id FROM rumah WHERE alamat = ?`, b.Alamat).Scan(&rumahID); err != nil {
			return h, err
		}

		var tahunLahir interface{}
		perkiraan := 0
		if b.UmurAda {
			tahunLahir, perkiraan = tahunImpor-b.Umur, 1
		}
		statusHuni := "pemilik"
		if b.Status == "kontrak" {
			statusHuni = "kontrak"
		}
		res, err = tx.Exec(`INSERT INTO penghuni (rumah_id, nama_kk, gender, tahun_lahir, tahun_lahir_perkiraan, jumlah_anggota, status_huni, mulai)
			SELECT ?, ?, ?, ?, ?, ?, ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM penghuni WHERE rumah_id = ? AND selesai IS NULL)`,
			rumahID, b.NamaPenghuni(), nullable(b.Gender), tahunLahir, perkiraan, b.JumlahAnggota, statusHuni, mulaiDefault, rumahID)
		if err != nil {
			return h, err
		}
		if n, _ := res.RowsAffected(); n > 0 {
			h.Penghuni++
		}
	}
	return h, tx.Commit()
}

func nullable(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
