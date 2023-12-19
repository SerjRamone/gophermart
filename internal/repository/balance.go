package repository

import (
	"context"
	"fmt"

	"github.com/SerjRamone/gophermart/internal/models"
	"github.com/jackc/pgx/v5"
)

// GetUserBalance ...
func (db *DB) GetUserBalance(ctx context.Context, userID string) (*models.UserBalance, error) {
	query := `
    SELECT
    	current_sum - withdrawn_sum AS current,
    	withdrawn_sum AS withdrawn
	FROM (
    	SELECT COALESCE(SUM(o.accrual), 0) AS current_sum
    	FROM "order" o 
    	WHERE o.user_id = $1
	) AS orders,
	(
    	SELECT COALESCE(SUM(w.total), 0) AS withdrawn_sum
    	FROM withdrawal w
    	WHERE w.user_id = $1
	) AS withdrawals;`
	row := db.pool.QueryRow(ctx, query, userID)
	ub := models.UserBalance{}
	err := row.Scan(&ub.Current, &ub.Withdrawn)
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("row scan error: %w", err)
		}
	}

	return &ub, nil
}

// CreateWithdrawal ...
// @todo check not unique number column value error
func (db *DB) CreateWithdrawal(ctx context.Context, userID string, number string, total float64) error {
	// check points balance
	row := db.pool.QueryRow(
		ctx,
		`SELECT COALESCE(SUM(accrual), 0) AS current FROM "order" WHERE user_id = $1 GROUP BY user_id;`,
		userID,
	)
	ub := models.UserBalance{}
	if err := row.Scan(&ub.Current); err != nil {
		return fmt.Errorf("row scan error: %w", err)
	}
	// to small points balance
	if ub.Current < total {
		return models.ErrNotEnoughPoints
	}

	// balance is ok
	_, err := db.pool.Exec(
		ctx,
		`INSERT INTO withdrawal (user_id, number, total) VALUES ($1, $2, $3);`,
		userID,
		number,
		total,
	)
	if err != nil {
		return fmt.Errorf("withdrawal insert error: %w", err)
	}
	return nil
}

// GetWithdrawals ...
func (db *DB) GetWithdrawals(ctx context.Context, userID string) ([]*models.Withdrawal, error) {
	var withdrwls []*models.Withdrawal
	rows, err := db.pool.Query(
		ctx,
		`SELECT total, number, created_at FROM withdrawal WHERE user_id = $1 ORDER BY created_at DESC;`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get withdrawals error: %w", err)
	}
	defer rows.Close()

	// scan query result
	for rows.Next() {
		var w models.Withdrawal
		if err := rows.Scan(&w.Total, &w.OrderNumber, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("rows scan error: %w", err)
		}

		withdrwls = append(withdrwls, &w)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows scan error: %w", err)
	}

	return withdrwls, nil
}
