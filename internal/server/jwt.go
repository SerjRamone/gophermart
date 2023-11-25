// Package server ...
package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/SerjRamone/gophermart/internal/models"
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

func isValid(token string, secret []byte) (bool, error) {
	claims := &Claims{}
	// check token algo method
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid signing alg method: %v", t.Header["alg"])
		}
		return secret, nil
	}

	// parse claims to struct
	t, err := jwt.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		return false, fmt.Errorf("token parse error: %w", err)
	}

	if !t.Valid {
		return false, fmt.Errorf("token is not valid: %w", err)
	}

	return true, nil
}

// getUserFromToken ...
func (bHandler *baseHandler) getUserFromToken(r *http.Request) (*models.User, error) {
	token := r.Header.Get("Authorization")
	claims := &Claims{}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		return bHandler.secret, nil
	}
	_, err := jwt.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("parsing token error: %w", err)
	}

	uf := models.UserForm{
		Login: claims.Login,
	}

	u, err := bHandler.storage.GetUser(r.Context(), uf)
	if err != nil {
		return nil, fmt.Errorf("get user from token error: %w", err)
	}

	return u, nil
}

// store user in context
// type key string
// const userContextKey key = "userKey"

// JwtMiddleware ...
func (bHandler baseHandler) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		isAuthorized, err := isValid(token, bHandler.secret)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !isAuthorized {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// store user in context
		// u, err := bHandler.getUserFromToken(r)
		// if err != nil {
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	logger.Error("failed to get user from token", zap.Error(err))
		// 	return
		// }
		// ctx := context.WithValue(r.Context(), userContextKey, u)
		// next.ServeHTTP(w, r.WithContext(ctx))

		next.ServeHTTP(w, r)
	})
}
