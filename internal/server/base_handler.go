// Package server ...
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/SerjRamone/gophermart/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// baseHandler base handler with storage inside
type baseHandler struct {
	secret    []byte
	tokenExpr int
	storage   Storage
	// ... etc
}

// Storage ...
type Storage interface {
	CreateUser(context.Context, models.UserForm) (*models.User, error)
	GetUser(context.Context, models.UserForm) (*models.User, error)
}

// NewBaseHandler creates new baseHandler
func NewBaseHandler(secret []byte, tokenExpr int, storage Storage) baseHandler {
	return baseHandler{
		secret:    secret,
		tokenExpr: tokenExpr,
		storage:   storage,
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

// getHash return hash from string
func (bHandler baseHandler) getHash(password string) (string, error) {
	// logger.Info("getting hash for password", zap.String("password", password))
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("get password hash error: %w", err)
	}
	return string(b), nil
}

// compareHashAndPass compares a bcrypt hashed password with its possible
// plaintext equivalent. Returns true on success, or false on failure.
func (bHandler baseHandler) compareHashAndPass(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}