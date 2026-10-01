package auth

import (
	"database/sql"
	"time"
)

// BuatUser membuat user aktif beserta satu peran. Digunakan createadmin dan
// (nanti) kelola user. Nomor WA dinormalisasi ke E.164.
func BuatUser(db *sql.DB, nama, noWA, peran string, gang int) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO user (nama, no_wa, aktif, created_at) VALUES (?, ?, 1, ?)`,
		nama, normalisasiWA(noWA), now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	if _, err := tx.Exec(`INSERT INTO user_peran (user_id, peran, gang, urutan) VALUES (?, ?, ?, 0)`, id, peran, gang); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}
