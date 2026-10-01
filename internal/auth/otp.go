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

// ErrThrottle muncul saat nomor melebihi batas minta OTP.
var ErrThrottle = errors.New("terlalu sering minta kode, coba beberapa saat lagi")

// ErrTidakValid muncul saat kode salah, kedaluwarsa, atau percobaan habis.
var ErrTidakValid = errors.New("kode salah atau kedaluwarsa")

const otpTTL = 5 * time.Minute
const otpMaksPercobaan = 5

// GenerateOTP membuat kode 6 digit, menyimpan hash-nya, dan mengembalikan
// kode polos untuk dikirim. Throttle: maks 3 kode aktif per nomor.
func GenerateOTP(db *sql.DB, noWA string) (string, error) {
	var n int
	batas := now().Add(-10 * time.Minute).Format(time.RFC3339)
	if err := db.QueryRow(`SELECT count(*) FROM otp WHERE no_wa = ? AND kedaluwarsa > ?`, noWA, batas).Scan(&n); err != nil {
		return "", err
	}
	if n >= 3 {
		return "", ErrThrottle
	}

	kode, err := kodeAcak(6)
	if err != nil {
		return "", err
	}
	kedaluwarsa := now().Add(otpTTL).Format(time.RFC3339)
	if _, err := db.Exec(`INSERT INTO otp (no_wa, kode_hash, kedaluwarsa, percobaan) VALUES (?, ?, ?, 0)`,
		noWA, hashOTP(noWA, kode), kedaluwarsa); err != nil {
		return "", err
	}
	return kode, nil
}

// VerifyOTP mencocokkan kode, menghitung percobaan, dan menghapus OTP bila benar.
func VerifyOTP(db *sql.DB, noWA, kode string) error {
	var hash, kedaluwarsa string
	var percobaan int
	err := db.QueryRow(`SELECT kode_hash, kedaluwarsa, percobaan FROM otp WHERE no_wa = ? AND kedaluwarsa > ? ORDER BY kedaluwarsa DESC LIMIT 1`,
		noWA, now().Format(time.RFC3339)).Scan(&hash, &kedaluwarsa, &percobaan)
	if err == sql.ErrNoRows {
		return ErrTidakValid
	}
	if err != nil {
		return err
	}
	if percobaan >= otpMaksPercobaan {
		return ErrTidakValid
	}
	if hash != hashOTP(noWA, kode) {
		_, _ = db.Exec(`UPDATE otp SET percobaan = percobaan + 1 WHERE no_wa = ? AND kode_hash = ?`, noWA, hash)
		return ErrTidakValid
	}
	_, _ = db.Exec(`DELETE FROM otp WHERE no_wa = ? AND kode_hash = ?`, noWA, hash)
	return nil
}

func hashOTP(noWA, kode string) string {
	h := sha256.Sum256([]byte(noWA + "|" + kode))
	return hex.EncodeToString(h[:])
}

func kodeAcak(panjang int) (string, error) {
	const digit = "0123456789"
	out := make([]byte, panjang)
	for i := range out {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digit))))
		if err != nil {
			return "", err
		}
		out[i] = digit[n.Int64()]
	}
	return string(out), nil
}

// now mengembalikan waktu WIB (Asia/Jakarta, tanpa DST).
func now() time.Time {
	return time.Now().In(time.FixedZone("WIB", 7*3600))
}
