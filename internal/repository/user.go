package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/SerjRamone/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueConstraintName = "user_login_key"

// CreateUser ...
func (db *DB) CreateUser(ctx context.Context, form models.UserForm) (*models.User, error) {
	row := db.pool.QueryRow(
		ctx,
		`INSERT INTO "user" (login, password) VALUES ($1, $2) RETURNING id, login, password;`,
		form.Login,
		form.Password,
	)
	u := models.User{}
	if err := row.Scan(&u.ID, &u.Login, &u.PasswordHash); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// check pg error for detect `duplicated login` error
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) && pgErr.ConstraintName == uniqueConstraintName {
				return nil, models.ErrUserAlreadyExists
			}
			return nil, fmt.Errorf("user with same login is already exists: %w", err)
		}
		return nil, fmt.Errorf("row scan error: %w", err)
	}

	return &u, nil
}

// GetUser ...
func (db *DB) GetUser(ctx context.Context, form models.UserForm) (*models.User, error) {
	row := db.pool.QueryRow(
		ctx,
		`SELECT id, login, password FROM "user" WHERE login = $1;`,
		form.Login,
	)
	u := models.User{}
	if err := row.Scan(&u.ID, &u.Login, &u.PasswordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotExists
		}
		return nil, fmt.Errorf("row scan error: %w", err)
	}

	return &u, nil
}
