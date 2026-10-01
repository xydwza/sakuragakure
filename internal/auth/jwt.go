package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

// TerbitkanToken membuat JWT untuk sebuah user (masa berlaku 180 hari).
func (a *Auth) TerbitkanToken(userID int64) (string, error) {
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(sesiTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(a.secret)
}

func (a *Auth) verifikasiToken(tokenStr string) (int64, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return a.secret, nil
	})
	if err != nil {
		return 0, err
	}
	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid {
		return 0, errors.New("token tidak valid")
	}
	return c.UserID, nil
}
