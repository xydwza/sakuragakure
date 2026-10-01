package format

import (
	"strings"
	"testing"

	"sakuragakure/internal/db"
)

func TestFormatRupiah(t *testing.T) {
	cases := map[int64]string{
		0:        "Rp 0",
		25000:    "Rp 25.000",
		1234567:  "Rp 1.234.567",
		2500000:  "Rp 2.500.000",
		-1000000: "-Rp 1.000.000",
	}
	for n, want := range cases {
		if got := FormatRupiah(n); got != want {
			t.Fatalf("FormatRupiah(%d) = %q, ingin %q", n, got, want)
		}
	}
}

func TestLabelBulan(t *testing.T) {
	if got := LabelBulan("2026-10"); got != "Oktober 2026" {
		t.Fatalf("LabelBulan = %q", got)
	}
	if got := BulanPendek("2026-04"); got != "Apr" {
		t.Fatalf("BulanPendek = %q", got)
	}
	if got := TglID("2026-04-05"); got != "5 Apr 2026" {
		t.Fatalf("TglID = %q", got)
	}
}

func TestChartSVG(t *testing.T) {
	rows := []db.GrafikKasBulananRow{
		{Bulan: "2026-09", Masuk: 100000, Keluar: 50000},
		{Bulan: "2026-10", Masuk: 200000, Keluar: 0},
	}
	svg := ChartSVG(rows)
	if !strings.Contains(svg, "Grafik pemasukan dan pengeluaran kas RT") {
		t.Fatalf("svg tanpa aria-label")
	}
	if !strings.Contains(svg, "<rect") {
		t.Fatalf("svg tanpa batang")
	}
	if strings.Count(svg, "<rect") != 4 {
		t.Fatalf("ingin 4 rect (2 bulan x 2 arah), dapat %d", strings.Count(svg, "<rect"))
	}
	if got := ChartSVG(nil); !strings.Contains(got, "Belum ada mutasi") {
		t.Fatalf("grafik kosong harusnya tampil empty state")
	}
}
