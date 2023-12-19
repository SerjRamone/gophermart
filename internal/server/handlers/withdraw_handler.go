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

type withdrawal struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

// Balance is "GET /api/user/balance/withdraw" handler
func (bHandler baseHandler) Withdraw(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// get user from token
	u, err := bHandler.getUserFromToken(r)
	if err != nil {
		logger.Error("get user from token error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	wd := withdrawal{}

	// read request body
	b, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error("read /withdraw request body error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// unmarshal body
	if err := json.Unmarshal(b, &wd); err != nil {
		logger.Error("unmarshal withdraw body error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// validate order number
	of := models.OrderForm{Number: wd.Order}
	if !of.IsValidNumber() {
		w.WriteHeader(http.StatusUnprocessableEntity) // 422
		return
	}

	// create withdrawal
	if err := bHandler.storage.CreateWithdrawal(ctx, u.ID, wd.Order, wd.Sum); err != nil {
		if errors.Is(err, models.ErrNotEnoughPoints) {
			w.WriteHeader(http.StatusPaymentRequired) // 402
			return
		}
		logger.Error("creaet withdrawal error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// Balance is "GET /api/user/balance/withdrawals" handler
func (bHandler baseHandler) Withdrawals(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// get user from token
	u, err := bHandler.getUserFromToken(r)
	if err != nil {
		logger.Error("get user from token error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// get models from storage
	withdrwls, err := bHandler.storage.GetWithdrawals(ctx, u.ID)
	if err != nil {
		logger.Error("get withdrawals list error", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check `no content`
	if len(withdrwls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// marshal list
	b, err := json.Marshal(&withdrwls)
	if err != nil {
		logger.Error("marshal withdrawals list error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// send response
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(b); err != nil {
		logger.Error("write response error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
