package auth

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"
)

const sesiTTL = 180 * 24 * time.Hour

// Auth memegang DB dan session manager untuk middleware.
type Auth struct {
	DB       *sql.DB
	Sessions *scs.SessionManager
}

// New menyiapkan session store (SQLite via modernc) dan manager scs.
func New(db *sql.DB, sessionKey []byte) *Auth {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		expiry INTEGER NOT NULL
	); CREATE INDEX IF NOT EXISTS sessions_expiry ON sessions(expiry)`); err != nil {
		log.Fatalf("gagal buat tabel sesi: %v", err)
	}

	sm := scs.New()
	sm.Store = &store{db: db}
	sm.Lifetime = sesiTTL
	sm.Cookie.Name = "sakuragakure_sesi"
	sm.Cookie.HttpOnly = true
	sm.Cookie.SameSite = http.SameSiteLaxMode
	sm.Cookie.Secure = os.Getenv("COOKIE_SECURE") == "true"
	sm.Cookie.Persist = true
	return &Auth{DB: db, Sessions: sm}
}

type store struct{ db *sql.DB }

func (s *store) Delete(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func (s *store) Find(token string) ([]byte, bool, error) {
	var b []byte
	var expiry int64
	err := s.db.QueryRow(`SELECT data, expiry FROM sessions WHERE token = ?`, token).Scan(&b, &expiry)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if expiry <= time.Now().Unix() {
		_ = s.Delete(token)
		return nil, false, nil
	}
	return b, true, nil
}

func (s *store) Commit(token string, b []byte, expiry time.Time) error {
	_, err := s.db.Exec(`INSERT INTO sessions (token, data, expiry) VALUES (?, ?, ?)
		ON CONFLICT(token) DO UPDATE SET data = excluded.data, expiry = excluded.expiry`,
		token, b, expiry.Unix())
	return err
}

// SessionKey membaca SESSION_KEY dari env atau memakai kunci dev (dengan peringatan).
func SessionKey() []byte {
	if k := os.Getenv("SESSION_KEY"); len(k) >= 32 {
		return []byte(k)
	}
	log.Println("PERINGATAN: SESSION_KEY tidak diatur, memakai kunci dev. Set SESSION_KEY (>=32 karakter) untuk produksi.")
	return []byte("sakuragakure-dev-session-key-32bytes!")
}
