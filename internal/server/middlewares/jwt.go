// Package middlewares ...
package middlewares

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is a custom JWT Claims Set
type Claims struct {
	Login string
	jwt.RegisteredClaims
}

// GenerateJWT ...
func GenerateJWT(secret []byte, login string, tokenExperation int) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(tokenExperation) * time.Second)),
		},
		Login: login,
	})

	token, err := t.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("token SignedString error: %w", err)
	}

	return token, nil
}
