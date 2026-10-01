package pages

import "sakuragakure/internal/db"

// ActionCard adalah satu kartu aksi di dasbor kelola.
type ActionCard struct {
	Label, Href string
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
