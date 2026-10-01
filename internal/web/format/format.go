package format

import (
	"strconv"
	"strings"

	"sakuragakure/internal/db"
)

var bulanPendek = []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
var bulanPanjang = []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

// FormatRupiah memformat integer rupiah gaya Indonesia: "Rp 1.234.567".
func FormatRupiah(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	if neg {
		return "-Rp " + b.String()
	}
	return "Rp " + b.String()
}

// LabelBulan mengubah "2026-10" menjadi "Oktober 2026".
func LabelBulan(periode string) string {
	if len(periode) < 7 {
		return periode
	}
	tahun := periode[:4]
	bulan, err := strconv.Atoi(periode[5:7])
	if err != nil || bulan < 1 || bulan > 12 {
		return periode
	}
	return bulanPanjang[bulan-1] + " " + tahun
}

// BulanPendek mengubah "2026-10" menjadi "Okt".
func BulanPendek(periode string) string {
	if len(periode) < 7 {
		return periode
	}
	bulan, err := strconv.Atoi(periode[5:7])
	if err != nil || bulan < 1 || bulan > 12 {
		return periode
	}
	return bulanPendek[bulan-1]
}

// TglID mengubah "2026-04-05" menjadi "5 Apr 2026".
func TglID(tanggal string) string {
	p := strings.Split(tanggal, "-")
	if len(p) != 3 {
		return tanggal
	}
	d, err1 := strconv.Atoi(p[2])
	m, err2 := strconv.Atoi(p[1])
	if err1 != nil || err2 != nil || m < 1 || m > 12 {
		return tanggal
	}
	return strconv.Itoa(d) + " " + bulanPendek[m-1] + " " + p[0]
}

// ChartSVG merender grafik batang pemasukan/pengeluaran kas RT (server-side).
func ChartSVG(rows []db.GrafikKasBulananRow) string {
	const W, H, pad = 640, 200, 28
	if len(rows) == 0 {
		return `<div class="empty">Belum ada mutasi kas.</div>`
	}
	max := int64(1)
	for _, r := range rows {
		if r.Masuk > max {
			max = r.Masuk
		}
		if r.Keluar > max {
			max = r.Keluar
		}
	}
	bw := (W - pad*2) / len(rows)
	var b strings.Builder
	b.WriteString(`<div class="chart"><svg viewBox="0 0 ` + strconv.Itoa(W) + ` ` + strconv.Itoa(H) + `" role="img" aria-label="Grafik pemasukan dan pengeluaran kas RT per bulan">`)
	b.WriteString(`<line x1="` + strconv.Itoa(pad) + `" x2="` + strconv.Itoa(W-pad) + `" y1="` + strconv.Itoa(H-24) + `" y2="` + strconv.Itoa(H-24) + `" stroke="var(--line)"/>`)
	for i, r := range rows {
		x := pad + i*bw
		hi := int(float64(r.Masuk) / float64(max) * float64(H-50))
		ho := int(float64(r.Keluar) / float64(max) * float64(H-50))
		b.WriteString(`<rect x="` + itoa(x+int(float64(bw)*0.18)) + `" y="` + itoa(H-24-hi) + `" width="` + itoa(int(float64(bw)*0.3)) + `" height="` + itoa(hi) + `" rx="4" fill="var(--leaf)"><title>Masuk ` + FormatRupiah(r.Masuk) + `</title></rect>`)
		b.WriteString(`<rect x="` + itoa(x+int(float64(bw)*0.52)) + `" y="` + itoa(H-24-ho) + `" width="` + itoa(int(float64(bw)*0.3)) + `" height="` + itoa(ho) + `" rx="4" fill="var(--clay)"><title>Keluar ` + FormatRupiah(r.Keluar) + `</title></rect>`)
		b.WriteString(`<text x="` + itoa(x+bw/2) + `" y="` + itoa(H-6) + `" text-anchor="middle">` + BulanPendek(r.Bulan) + `</text>`)
	}
	b.WriteString(`</svg><div class="legend"><span><i style="background:var(--leaf);border-color:var(--leaf)"></i>Pemasukan iuran</span><span><i style="background:var(--clay);border-color:var(--clay)"></i>Pengeluaran</span></div></div>`)
	return b.String()
}

func itoa(n int) string { return strconv.Itoa(n) }
