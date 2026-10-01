package audit

import (
	"database/sql"
	"time"
)

// Tulis mencatat satu aksi ke audit_log.
func Tulis(db *sql.DB, userID int64, aksi, objek, objekID, detail, ip string) {
	var uid interface{}
	if userID != 0 {
		uid = userID
	}
	_, _ = db.Exec(`INSERT INTO audit_log (user_id, aksi, objek, objek_id, detail, ip, at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		uid, aksi, objek, objekID, detail, ip, time.Now().Format(time.RFC3339))
}
