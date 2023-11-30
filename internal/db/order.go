package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/SerjRamone/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// CreateOrder ...
func (db *DB) CreateOrder(ctx context.Context, form models.OrderForm) (*models.Order, error) {
	// @todo set accrual as 0 by default. maybe use DEFAULT on CREATE TABLE query?
	row := db.pool.QueryRow(
		ctx,
		`INSERT INTO "order" (user_id, number, status, accrual) VALUES ($1, $2, $3, 0) RETURNING id, user_id, number, accrual, status, uploaded_at;`,
		form.UserID,
		form.Number,
		models.OrderStatusNew,
	)
	o := models.Order{}
	if err := row.Scan(&o.ID, &o.UserID, &o.Number, &o.Accrual, &o.Status, &o.UploadedAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// check pg error for detect `duplicated number` error
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) && pgErr.ConstraintName == "order_number_key" {
				return nil, models.ErrOrderAlreadyExists
			}
			return nil, fmt.Errorf("order with same number is already exists: %w", err)
		}
		return nil, fmt.Errorf("row scan error: %w", err)
	}

	return &o, nil
}

// GetOrder ...
func (db *DB) GetOrder(ctx context.Context, form models.OrderForm) (*models.Order, error) {
	row := db.pool.QueryRow(
		ctx,
		`SELECT id, user_id, number, status, accrual, uploaded_at FROM "order" WHERE number = $1;`,
		form.Number,
	)
	o := models.Order{}
	if err := row.Scan(&o.ID, &o.UserID, &o.Number, &o.Status, &o.Accrual, &o.UploadedAt); err != nil {
		return nil, fmt.Errorf("row scan error: %w", err)
	}

	return &o, nil
}

// GetUserOrders ...
func (db *DB) GetUserOrders(ctx context.Context, u *models.User) ([]*models.Order, error) {
	rows, err := db.pool.Query(
		ctx,
		`SELECT id, number, accrual, status, uploaded_at FROM "order" WHERE user_id = $1 ORDER BY uploaded_at ASC;`,
		u.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("get data from storrage error: %w", err)
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.Number, &o.Accrual, &o.Status, &o.UploadedAt); err != nil {
			return nil, fmt.Errorf("row scan erroro: %w", err)
		}
		orders = append(orders, &o)
	}

	return orders, nil
}

// GetUnprocessedOrders ...
func (db *DB) GetUnprocessedOrders(ctx context.Context) ([]*models.Order, error) {
	// do query
	rows, err := db.pool.Query(
		ctx,
		`SELECT id, user_id, number, accrual, status, uploaded_at FROM "order" WHERE status IN ($1, $2);`,
		models.OrderStatusNew,
		models.OrderStatusProcessing,
	)
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	defer rows.Close()

	// scan rows into slice
	var orders []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Number, &o.Accrual, &o.Status, &o.UploadedAt); err != nil {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
		orders = append(orders, &o)
	}

	return orders, nil
}

// UpdateOrder ...
func (db *DB) UpdateOrder(ctx context.Context, order *models.Order) error {
	_, err := db.pool.Exec(
		ctx,
		`UPDATE "order" SET accrual = $2, status = $3 WHERE id = $1;`,
		order.ID,
		order.Accrual,
		order.Status,
	)
	if err != nil {
		return fmt.Errorf("order update error: %w", err)
	}

	return nil
}
