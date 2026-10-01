package auth

import (
	"database/sql"
	"time"
)

// BuatUser membuat user aktif dengan kata sandi (ter-hash) beserta satu peran.
func BuatUser(db *sql.DB, nama, noWA, peran, password string, gang int) (int64, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return 0, err
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO user (nama, no_wa, aktif, password_hash, created_at) VALUES (?, ?, 1, ?, ?)`,
		nama, normalisasiWA(noWA), hash, now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	if _, err := tx.Exec(`INSERT INTO user_peran (user_id, peran, gang, urutan) VALUES (?, ?, ?, 0)`, id, peran, gang); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}
