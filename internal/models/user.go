// Package models ...
package models

import (
	"errors"
	// "github.com/google/uuid"
)

var (
	// ErrUserAlreadyExists is not unique login error
	ErrUserAlreadyExists = errors.New("login is already exists")

	// ErrUserNotExists user not found error
	ErrUserNotExists = errors.New("user is not exists")
)

// UserForm data object from request
type UserForm struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// User data object from storage
type User struct {
	ID string `json:"id"`
	// ID           uuid.UUID `json:"id"`
	Login        string `json:"login"`
	PasswordHash string `json:"password"`
	// CreatedAt
}

// UserBalance current accrualed balance and total withdrawned
type UserBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
