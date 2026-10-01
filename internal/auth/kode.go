package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"math/big"
	"time"
)

// ErrKodeKedaluwarsa / ErrKodeDipakai / ErrKodeSalah untuk kode cadangan.
var (
	ErrKodeSalah      = errors.New("kode cadangan salah")
	ErrKodeKedaluwarsa = errors.New("kode cadangan kedaluwarsa")
)

const kodeTTL = 24 * time.Hour

const kodeAbjad = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// BuatKodeCadangan membuat kode sekali pakai (24 jam) untuk user.
func BuatKodeCadangan(db *sql.DB, userID, dibuatOleh int64) (string, error) {
	kode, err := kodeAcakAbjad(8)
	if err != nil {
		return "", err
	}
	kedaluwarsa := now().Add(kodeTTL).Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO kode_cadangan (kode_hash, user_id, dibuat_oleh, kedaluwarsa) VALUES (?, ?, ?, ?)`,
		hashKode(kode), userID, dibuatOleh, kedaluwarsa); err != nil {
		return "", err
	}
	return kode, nil
}

// PakaiKodeCadangan memakai kode untuk login; mengembalikan user_id bila sah.
func PakaiKodeCadangan(db *sql.DB, kode string) (int64, error) {
	var userID int64
	var kedaluwarsa string
	var dipakai sql.NullString
	err := db.QueryRow(`SELECT user_id, kedaluwarsa, dipakai_at FROM kode_cadangan WHERE kode_hash = ?`, hashKode(kode)).
		Scan(&userID, &kedaluwarsa, &dipakai)
	if err == sql.ErrNoRows {
		return 0, ErrKodeSalah
	}
	if err != nil {
		return 0, err
	}
	if dipakai.Valid {
		return 0, ErrKodeSalah
	}
	if kedaluwarsa <= now().Format(time.RFC3339) {
		return 0, ErrKodeKedaluwarsa
	}
	if _, err := db.Exec(`UPDATE kode_cadangan SET dipakai_at = ? WHERE kode_hash = ?`, now().Format(time.RFC3339), hashKode(kode)); err != nil {
		return 0, err
	}
	return userID, nil
}

func hashKode(kode string) string {
	h := sha256.Sum256([]byte(kode))
	return hex.EncodeToString(h[:])
}

func kodeAcakAbjad(panjang int) (string, error) {
	out := make([]byte, panjang)
	for i := range out {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(kodeAbjad))))
		if err != nil {
			return "", err
		}
		out[i] = kodeAbjad[n.Int64()]
	}
	return string(out), nil
}
