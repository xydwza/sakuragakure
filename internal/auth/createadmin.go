package auth

import (
	"database/sql"
	"strings"
	"time"
)

// BuatUser membuat user aktif dengan username + kata sandi (ter-hash) beserta satu peran.
// Bila password kosong, dipakai username.
func BuatUser(db *sql.DB, nama, username, noWA, peran, password string, gang int) (int64, error) {
	if password == "" {
		password = username
	}
	hash, err := hashPassword(password)
	if err != nil {
		return 0, err
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO user (nama, username, no_wa, aktif, password_hash, created_at) VALUES (?, ?, ?, 1, ?, ?)`,
		nama, strings.ToLower(username), NormalisasiWA(noWA), hash, now().Format(time.RFC3339))
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	if _, err := tx.Exec(`INSERT INTO user_peran (user_id, peran, gang, urutan) VALUES (?, ?, ?, 0)`, id, peran, gang); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}
