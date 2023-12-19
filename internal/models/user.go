// Package models ...
package models

import (
	"errors"
	"time"
	// "github.com/google/uuid"
)

var (
	// ErrUserAlreadyExists is not unique login error
	ErrUserAlreadyExists = errors.New("login is already exists")

	// ErrUserNotExists user not found error
	ErrUserNotExists = errors.New("user is not exists")

	// ErrNotEnoughPoints too small points balance error
	ErrNotEnoughPoints = errors.New("not enough points")
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

// Withdrawal some point withdrawal operation
type Withdrawal struct {
	OrderNumber string    `json:"order"`
	Total       float64   `json:"sum"`
	CreatedAt   time.Time `json:"processed_at"`
}
