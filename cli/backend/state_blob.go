package backend

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const stateTTL = 10 * time.Minute

type stateBlob struct {
	Nonce    string `json:"nonce"`
	Verifier string `json:"verifier"`
	Return   string `json:"return"`
	jwt.RegisteredClaims
}

func encodeState(secret []byte, blob stateBlob) (string, error) {
	now := time.Now()
	blob.IssuedAt = jwt.NewNumericDate(now)
	blob.ExpiresAt = jwt.NewNumericDate(now.Add(stateTTL))
	return jwt.NewWithClaims(jwt.SigningMethodHS256, blob).SignedString(secret)
}

func decodeState(secret []byte, raw string) (stateBlob, error) {
	blob := stateBlob{}
	token, err := jwt.ParseWithClaims(raw, &blob,
		func(*jwt.Token) (interface{}, error) { return secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return stateBlob{}, err
	}
	if !token.Valid || blob.ID == "" || blob.Nonce == "" || blob.Verifier == "" {
		return stateBlob{}, errors.New("incomplete state")
	}
	return blob, nil
}
