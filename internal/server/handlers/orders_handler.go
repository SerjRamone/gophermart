package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/SerjRamone/gophermart/internal/models"
	"github.com/SerjRamone/gophermart/pkg/logger"
	"go.uber.org/zap"
)

// PostOrder is "POST /api/user/orders" handler
func (bHandler baseHandler) PostOrder(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// get user from token
	u, err := bHandler.getUserFromToken(r)
	if err != nil {
		logger.Error("get user from token error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}

	// read request body
	b, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("reading request body error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError) // 500
		return
	}
	if len(b) == 0 {
		logger.Error("empty body")
		w.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	of := models.OrderForm{
		Number: string(b),
		UserID: u.ID,
	}

	// validate order number
	if !of.IsValidNumber() {
		w.WriteHeader(http.StatusUnprocessableEntity) // 422
		return
	}

	// create order in storage
	if _, err = bHandler.storage.CreateOrder(ctx, of); err != nil {
		// get unknown error
		if !errors.Is(err, models.ErrOrderAlreadyExists) {
			logger.Error("order create error", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// try to get order from storage
		o, err := bHandler.storage.GetOrder(ctx, of)
		if err != nil {
			logger.Error("order get error", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		// check if order by another user
		if o.UserID != u.ID {
			w.WriteHeader(http.StatusConflict) // 409
			return
		}

		w.WriteHeader(http.StatusOK) // 200
		return
	}

	w.WriteHeader(http.StatusAccepted) // 202
}

// GetOrder is "GET /api/user/orders" handler
func (bHandler baseHandler) GetOrder(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// get user from token
	u, err := bHandler.getUserFromToken(r)
	if err != nil {
		logger.Error("get user from token error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// get orders from storage
	orders, err := bHandler.storage.GetUserOrders(ctx, u)
	if err != nil {
		logger.Error("get user's order error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// empty response
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent) // 204
		return
	}

	// marshal orders
	b, err := json.Marshal(orders)
	if err != nil {
		logger.Error("marshal ordders error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write response
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(b); err != nil {
		logger.Error("write response error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
