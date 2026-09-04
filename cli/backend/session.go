package backend

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const SessionTTL = 12 * time.Hour

type Session struct {
	Username string `json:"usr"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}

func SignSession(secret []byte, session Session, ttl time.Duration) (string, error) {
	now := time.Now()
	session.IssuedAt = jwt.NewNumericDate(now)
	session.ExpiresAt = jwt.NewNumericDate(now.Add(ttl))
	return jwt.NewWithClaims(jwt.SigningMethodHS256, session).SignedString(secret)
}

func VerifySession(secret []byte, raw string) (Session, error) {
	session := Session{}
	token, err := jwt.ParseWithClaims(raw, &session,
		func(*jwt.Token) (interface{}, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return Session{}, err
	}
	if !token.Valid {
		return Session{}, errors.New("invalid session")
	}
	return session, nil
}
