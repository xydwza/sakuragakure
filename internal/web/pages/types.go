package pages

import (
	"database/sql"

	"sakuragakure/internal/db"
)

// ActionCard adalah satu kartu aksi di dasbor kelola.
type ActionCard struct {
	Label, Href string
}

// NodePengurus adalah simpul pohon struktur pengurus.
type NodePengurus struct {
	Nama, Jabatan, Wa, Alamat, Detail string
	Peran                             string
	Gang                              int
	FotoMediaID                       sql.NullInt64
	Anak                              []NodePengurus
}

// BuildPohon menyusun daftar pengurus menjadi pohon (ketua -> inti & koordinator -> pembantu).
func BuildPohon(list []db.ListPengurusRow) NodePengurus {
	var ketua NodePengurus
	var inti []NodePengurus
	koord := map[int64]*NodePengurus{}
	pembantu := map[int64][]NodePengurus{}

	toNode := func(p db.ListPengurusRow) NodePengurus {
		return NodePengurus{
			Nama:        p.Nama,
			Jabatan:     PeranLabel(p.Peran, int(p.Gang)),
			Wa:          p.NoWa,
			Alamat:      p.Alamat.String,
			Detail:      p.Detail.String,
			Peran:       p.Peran,
			Gang:        int(p.Gang),
			FotoMediaID: p.FotoMediaID,
		}
	}
	for _, p := range list {
		switch p.Peran {
		case "ketua":
			ketua = toNode(p)
		case "wakil", "sekretaris", "bendahara":
			inti = append(inti, toNode(p))
		case "koordinator":
			n := toNode(p)
			koord[p.Gang] = &n
		case "pembantu_koordinator":
			pembantu[p.Gang] = append(pembantu[p.Gang], toNode(p))
		}
	}
	ketua.Anak = append(ketua.Anak, inti...)
	for g := int64(1); g <= 3; g++ {
		if k, ok := koord[g]; ok {
			k.Anak = append(k.Anak, pembantu[g]...)
			ketua.Anak = append(ketua.Anak, *k)
		}
	}
	return ketua
}

// BlokRumah adalah sekelompok rumah dalam satu blok untuk grid tarik iuran.
type BlokRumah struct {
	Blok  string
	Rumah []db.GridGangRow
}

// TileStatus menghitung kelas visual kotak rumah.
func TileClass(r db.GridGangRow, locked bool) string {
	switch {
	case r.Status == "kosong":
		return "kosong"
	case r.BebasIuranAlasan.Valid:
		return "bebas"
	case r.BayarStatus == "diterima":
		return "done"
	case r.BayarStatus == "disetor":
		return "held locked"
	case r.BayarStatus == "dipegang":
		return "held"
	}
	if locked {
		return "locked"
	}
	return ""
}

// TileMark mengembalikan tanda di pojok kotak.
func TileMark(r db.GridGangRow) string {
	switch r.BayarStatus {
	case "diterima", "dipegang":
		return "✓"
	case "disetor":
		return "↑"
	}
	return ""
}

// TileSub adalah keterangan kecil di bawah nomor rumah.
func TileSub(r db.GridGangRow) string {
	switch {
	case r.Status == "kosong":
		return "Kosong"
	case r.BebasIuranAlasan.Valid:
		return "Bebas: " + r.BebasIuranAlasan.String
	case r.NamaKk != "":
		return r.NamaKk
	}
	return "Penghuni belum diisi"
}
