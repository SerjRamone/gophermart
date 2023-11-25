package models

import (
	"errors"
	"strconv"
	"time"
)

var (
	// ErrOrderAlreadyExists is not unique order number error
	ErrOrderAlreadyExists = errors.New("order is already exists")
)

// order statuses
const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

// OrderForm data object from request
type OrderForm struct {
	UserID string `json:"user_id"`
	Number string `json:"number"`
}

// Order data object from storage
type Order struct {
	ID         string    `json:"order_id"`
	UserID     string    `json:"user_id,omitempty"`
	Status     string    `json:"status"`
	Number     string    `json:"number"`
	Accrual    float64   `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// IsValidNumber returns true if OrderForm.Number checked by Luhn algorithm
func (of OrderForm) IsValidNumber() bool {
	sum := 0
	double := false
	for i := len(of.Number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(of.Number[i]))
		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}
	return sum%10 == 0
}
