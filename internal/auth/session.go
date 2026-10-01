package auth

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"
)

const tokenCookie = "token"
const sesiTTL = 180 * 24 * time.Hour

// Auth memegang DB dan kunci JWT untuk otentikasi.
type Auth struct {
	DB     *sql.DB
	secret []byte
}

func New(conn *sql.DB, secret []byte) *Auth {
	return &Auth{DB: conn, secret: secret}
}

// SecretKey membaca SESSION_KEY dari env atau memakai kunci dev.
func SecretKey() []byte {
	if k := os.Getenv("SESSION_KEY"); len(k) >= 32 {
		return []byte(k)
	}
	log.Println("PERINGATAN: SESSION_KEY tidak diatur, memakai kunci dev.")
	return []byte("sakuragakure-dev-session-key-32bytes!")
}

type ctxKey struct{}

// Middleware memverifikasi JWT dari cookie dan menaruh userID di konteks.
func (a *Auth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if c, err := r.Cookie(tokenCookie); err == nil {
			if uid, err := a.verifikasiToken(c.Value); err == nil {
				ctx = context.WithValue(ctx, ctxKey{}, uid)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserID mengembalikan id user dari konteks (0 bila belum masuk).
func (a *Auth) UserID(ctx context.Context) int64 {
	v, _ := ctx.Value(ctxKey{}).(int64)
	return v
}

// cookieSecure mengikuti env COOKIE_SECURE.
func cookieSecure() bool { return os.Getenv("COOKIE_SECURE") == "true" }

// now mengembalikan waktu WIB (Asia/Jakarta, tanpa DST).
func now() time.Time {
	return time.Now().In(time.FixedZone("WIB", 7*3600))
}
