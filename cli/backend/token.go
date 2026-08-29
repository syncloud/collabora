package backend

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTokenTTL = 10 * time.Hour

type AccessToken struct {
	FileID string `json:"fid"`
	User   string `json:"usr"`
	jwt.RegisteredClaims
}

func SignAccessToken(secret []byte, fileID, user string, ttl time.Duration) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(jwt.SigningMethodHS256, AccessToken{
		FileID: fileID,
		User:   user,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}).SignedString(secret)
}

func VerifyAccessToken(secret []byte, raw, fileID string) (AccessToken, error) {
	claims := AccessToken{}
	token, err := jwt.ParseWithClaims(raw, &claims,
		func(*jwt.Token) (interface{}, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return AccessToken{}, err
	}
	if !token.Valid {
		return AccessToken{}, errors.New("invalid access token")
	}
	if claims.FileID != fileID {
		return AccessToken{}, errors.New("access token file mismatch")
	}
	return claims, nil
}
