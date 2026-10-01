package notif

import (
	"database/sql"
	"time"
)

// Enqueue menambahkan pesan WA ke antrian outbox (dikirim worker di fase 1j).
func Enqueue(db *sql.DB, noWA, template, payload string) error {
	_, err := db.Exec(`INSERT INTO notif_outbox (no_wa, template, payload, status, percobaan, kirim_setelah)
		VALUES (?, ?, ?, 'antri', 0, ?)`,
		noWA, template, payload, time.Now().Format(time.RFC3339))
	return err
}
