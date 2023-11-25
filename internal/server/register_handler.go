package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/SerjRamone/gophermart/internal/models"
	"github.com/SerjRamone/gophermart/pkg/logger"
	"go.uber.org/zap"
)

// Register handles user registrations
func (bHandler baseHandler) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// get UserForm from request
	uf, err := bHandler.getCredentials(w, r)
	if err != nil {
		logger.Error("get credentails error", zap.Error(err))
		return
	}

	// hash password
	uf.Password, err = bHandler.getHash(uf.Password)
	if err != nil {
		logger.Error("get password hash error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// create user
	if _, err = bHandler.storage.CreateUser(ctx, *uf); err != nil {
		if errors.Is(err, models.ErrUserAlreadyExists) {
			logger.Error("login is already exists", zap.Error(err))
			w.WriteHeader(http.StatusConflict)
			return
		}

		logger.Error("failed add user in the /register request", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// w.WriteHeader(http.StatusInternalServerError)
	// return

	// generate auth token
	token, err := GenerateJWT(bHandler.secret, uf.Login, bHandler.tokenExpr)
	if err != nil {
		logger.Error("generate JWT error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}
