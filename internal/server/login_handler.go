package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/SerjRamone/gophermart/internal/models"
	"github.com/SerjRamone/gophermart/pkg/logger"
	"go.uber.org/zap"
)

// Login handles user authorization
func (bHandler baseHandler) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// get credentials from request
	uf, err := bHandler.getCredentials(w, r)
	if err != nil {
		logger.Error("failed to get credentials", zap.Error(err))
		return
	}

	// find user in storage
	u, err := bHandler.storage.GetUser(ctx, *uf)
	if err != nil {
		if errors.Is(err, models.ErrUserNotExists) {
			logger.Error("user not found", zap.Error(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		logger.Error("get user error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// hashes did not match
	if !bHandler.compareHashAndPass(u.PasswordHash, uf.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := GenerateJWT(bHandler.secret, u.Login, bHandler.tokenExpr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}
