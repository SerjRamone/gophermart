package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/SerjRamone/gophermart/pkg/logger"
	"go.uber.org/zap"
)

// Balance is "GET /api/user/balance" handler
func (bHandler baseHandler) Balance(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// get user from token
	u, err := bHandler.getUserFromToken(r)
	if err != nil {
		logger.Error("get user from token error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	balance, err := bHandler.storage.GetUserBalance(ctx, u.ID)
	if err != nil {
		logger.Error("get users's balance error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// marshal orders
	b, err := json.Marshal(balance)
	if err != nil {
		logger.Error("marshal balance error", zap.Error(err))
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
