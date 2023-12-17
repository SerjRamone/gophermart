// Package repository contents common code for working with database
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"time"

	"github.com/SerjRamone/gophermart/internal/server/handlers"
	"github.com/SerjRamone/gophermart/migrations"
	"github.com/SerjRamone/gophermart/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/pressly/goose/v3"
)

var _ handlers.Storage = (*DB)(nil)

// DB ...
type DB struct {
	pool *pgxpool.Pool
}

// NewDB creates DB instance
func NewDB(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to created pool: %w", err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquiring connection from pool error: %w", err)
	}
	defer conn.Release()

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err = conn.Ping(pingCtx)
	if err != nil {
		return nil, fmt.Errorf("could not get pong from db: %w", err)
	}

	logger.Info("connected to the DB")

	db := stdlib.OpenDBFromPool(pool)
	if err := applyMigrations(db, migrations.SQLFiles); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	defer db.Close()

	return &DB{
		pool: pool,
	}, nil
}

// Close closes all connections in the pool
func (db *DB) Close() {
	logger.Info("close db connections")
	db.pool.Close()
}

// applyMigrations applies DB migrations (goose.Up())
func applyMigrations(db *sql.DB, fsys fs.FS) error {
	goose.SetBaseFS(fsys)
	goose.SetSequential(true)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose.Setdialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose.Up: %w", err)
	}
	return nil
}
