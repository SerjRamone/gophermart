package db

import (
	"context"
	"fmt"

	"github.com/SerjRamone/gophermart/internal/models"
)

// GetUserBalance ...
func (db *DB) GetUserBalance(ctx context.Context, userID string) (*models.UserBalance, error) {
	query := `
    SELECT 
		COALESCE(SUM(accrual), 0) AS current, 
        COALESCE(SUM(total), 0) AS withdrawn
    FROM "order"
    LEFT JOIN withdrawl ON "order".user_id = withdrawl.user_id
    WHERE "order".user_id = $1
    GROUP BY "order".user_id, withdrawl.user_id;`
	row := db.pool.QueryRow(ctx, query, userID)
	ub := models.UserBalance{}
	if err := row.Scan(&ub.Current, &ub.Withdrawn); err != nil {
		return nil, fmt.Errorf("row scan error: %w", err)
	}

	return &ub, nil
}
