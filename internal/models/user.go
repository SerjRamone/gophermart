// Package models ...
package models

import (
	"context"
	"errors"
	"fmt"
	// "github.com/google/uuid"
)

// var _ UserStorage = (*UserForm)(nil)

var (
	// ErrAlreadyExists is not unique login error
	ErrAlreadyExists = errors.New("login is already exists")

	// ErrNotExists user not found error
	ErrNotExists = errors.New("user is not exists")
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

// UserStorage ...
type UserStorage interface {
	CreateUser(context.Context, UserForm) (*User, error)
	GetUser(context.Context, UserForm) (*User, error)
}

// GetUser ...
func (uf UserForm) GetUser(ctx context.Context, storage UserStorage) (*User, error) {
	u, err := storage.GetUser(ctx, uf)
	if err != nil {
		return nil, fmt.Errorf("user get error: %w", err)
	}
	return u, nil
}

// CreateUser ...
func (uf UserForm) CreateUser(ctx context.Context, storage UserStorage) (*User, error) {
	u, err := storage.CreateUser(ctx, uf)
	if err != nil {
		return nil, fmt.Errorf("user add error: %w", err)
	}
	return u, nil
}
