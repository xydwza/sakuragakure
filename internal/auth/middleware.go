package auth

import (
	"context"
	"net/http"
)

// Peran adalah satu peran user (dengan gang bila terikat).
type Peran struct {
	Peran string
	Gang  int
}

// UserID mengembalikan id user dari sesi (0 bila belum masuk).
func (a *Auth) UserID(ctx context.Context) int64 {
	return a.Sessions.GetInt64(ctx, "userID")
}

// PeranList mengembalikan semua peran user yang sedang masuk.
func (a *Auth) PeranList(ctx context.Context) ([]Peran, error) {
	uid := a.UserID(ctx)
	if uid == 0 {
		return nil, nil
	}
	rows, err := a.DB.Query(`SELECT peran, gang FROM user_peran WHERE user_id = ?`, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Peran
	for rows.Next() {
		var p Peran
		if err := rows.Scan(&p.Peran, &p.Gang); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// RequireLogin mengalihkan tamu ke /masuk.
func (a *Auth) RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.UserID(r.Context()) == 0 {
			http.Redirect(w, r, "/masuk", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequirePeran membolehkan hanya peran tertentu; tamu dialihkan, peran salah 403.
func (a *Auth) RequirePeran(peran ...string) func(http.Handler) http.Handler {
	set := make(map[string]bool, len(peran))
	for _, p := range peran {
		set[p] = true
	}
	return func(next http.Handler) http.Handler {
		return a.RequireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			list, err := a.PeranList(r.Context())
			if err != nil {
				http.Error(w, "kesalahan internal", http.StatusInternalServerError)
				return
			}
			for _, p := range list {
				if set[p.Peran] {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "tidak berhak", http.StatusForbidden)
		}))
	}
}
