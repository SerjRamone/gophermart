// Package security ...
package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type hasher struct{}

// NewHasher ...
func NewHasher() *hasher {
	return &hasher{}
}

// GetHash returns hash from string
func (h *hasher) GetHash(password string) (string, error) {
	// logger.Info("getting hash for password", zap.String("password", password))
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("get password hash error: %w", err)
	}
	return string(b), nil
}

// CompareHashAndPass compares a bcrypt hashed password with its possible
// plaintext equivalent. Returns true on success, or false on failure.
func (h *hasher) CompareHashAndPass(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
