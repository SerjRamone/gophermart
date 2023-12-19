// Package handlers ...
//
//go:generate gotests --all -w base_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/SerjRamone/gophermart/internal/models"
	"github.com/SerjRamone/gophermart/internal/server/middlewares"
	"github.com/golang-jwt/jwt/v5"
)

// baseHandler base handler with storage inside
type baseHandler struct {
	secret    []byte
	tokenExpr int
	storage   Storage
	hasher    Hasher
	// ... etc
}

// Storage ...
type Storage interface {
	CreateUser(context.Context, models.UserForm) (*models.User, error)
	GetUser(context.Context, models.UserForm) (*models.User, error)
	CreateOrder(context.Context, models.OrderForm) (*models.Order, error)
	GetOrder(context.Context, models.OrderForm) (*models.Order, error)
	GetUserOrders(ctx context.Context, order *models.User) ([]*models.Order, error)
	GetUserBalance(ctx context.Context, userID string) (*models.UserBalance, error)
	CreateWithdrawal(ctx context.Context, userID string, number string, total float64) error
	GetWithdrawals(ctx context.Context, userID string) ([]*models.Withdrawal, error)
	GetUnprocessedOrders(ctx context.Context) ([]*models.Order, error)
	UpdateOrder(ctx context.Context, order *models.Order) error
}

// Hasher ...
type Hasher interface {
	GetHash(password string) (string, error)
	CompareHashAndPass(hash, password string) bool
}

// NewBaseHandler creates new baseHandler
func NewBaseHandler(secret []byte, tokenExpr int, storage Storage, hasher Hasher) baseHandler {
	return baseHandler{
		secret:    secret,
		tokenExpr: tokenExpr,
		storage:   storage,
		hasher:    hasher,
	}
}

// getCredentials return UserFrom model from request or error
func (bHandler baseHandler) getCredentials(w http.ResponseWriter, r *http.Request) (*models.UserForm, error) {
	var u models.UserForm

	// read requst body
	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("request body read error: %w", err)
	}

	// unmarshal body with credentials
	if err := json.Unmarshal(b, &u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("unmarshalling error: %w", err)
	}

	// login is required
	if u.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("login is empty")
	}

	return &u, nil
}

// getUserFromToken ...
func (bHandler *baseHandler) getUserFromToken(r *http.Request) (*models.User, error) {
	token := r.Header.Get("Authorization")
	claims := &middlewares.Claims{}

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

func isValid(token string, secret []byte) (bool, error) {
	claims := &middlewares.Claims{}
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
