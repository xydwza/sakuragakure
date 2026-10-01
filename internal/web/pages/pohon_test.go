package pages

import (
	"testing"

	"sakuragakure/internal/db"
)

func TestBuildPohon(t *testing.T) {
	list := []db.ListPengurusRow{
		{Nama: "Jepri", Peran: "ketua"},
		{Nama: "Wiyanto", Peran: "wakil"},
		{Nama: "Ferry", Peran: "sekretaris"},
		{Nama: "Danu", Peran: "bendahara"},
		{Nama: "Ipan", Peran: "koordinator", Gang: 1},
		{Nama: "Ajat", Peran: "koordinator", Gang: 2},
		{Nama: "Wahyu", Peran: "pembantu_koordinator", Gang: 2},
		{Nama: "Bambang", Peran: "koordinator", Gang: 3},
		{Nama: "Sagito", Peran: "pembantu_koordinator", Gang: 3},
	}

	pohon := BuildPohon(list)
	if pohon.Nama != "Jepri" {
		t.Fatalf("akar harusnya Jepri, dapat %s", pohon.Nama)
	}
	// anak ketua = inti (3) + koordinator (3) = 6
	if len(pohon.Anak) != 6 {
		t.Fatalf("anak ketua harus 6, dapat %d", len(pohon.Anak))
	}
	// koordinator gang 2 punya 1 pembantu
	var koord2 NodePengurus
	for _, a := range pohon.Anak {
		if a.Peran == "koordinator" && a.Gang == 2 {
			koord2 = a
		}
	}
	if len(koord2.Anak) != 1 || koord2.Anak[0].Nama != "Wahyu" {
		t.Fatalf("koordinator gang 2 harusnya 1 pembantu (Wahyu)")
	}
}
