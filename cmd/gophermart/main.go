// Package main ...
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/SerjRamone/gophermart/internal/accrual"
	"github.com/SerjRamone/gophermart/internal/config"
	"github.com/SerjRamone/gophermart/internal/repository"
	"github.com/SerjRamone/gophermart/internal/server/router"
	"github.com/SerjRamone/gophermart/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// parse config
	conf, err := config.NewGophermart()
	if err != nil {
		return err
	}

	// init logger
	if err := logger.New(
		logger.LevelOption(conf.LogLevel),
		logger.TimeKeyOption("timestamp"),
		logger.TimeEncoderOption(zapcore.ISO8601TimeEncoder),
	); err != nil {
		return err
	}

	logger.Info("loaded config", zap.Object("config", &conf))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	db, err := repository.NewDB(ctx, conf.DatabaseURI)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:    conf.RunAddress,
		Handler: router.NewRouter([]byte(conf.SecretKey), conf.TokenExpiration, db),
	}

	go func() {
		logger.Info("starting server...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server start error", zap.Error(err))
			cancel()
		}
	}()

	// @todo
	// start watching orders
	// new accrual client
	accrualClient := accrual.NewAccrualClient(conf.AccrualSystemAddress)
	accrualClient.WatchOrders(ctx, db)

	<-ctx.Done()

	// shutting down server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutting down error", zap.Error(err))
	} else {
		logger.Info("server shut down gracefully")
	}

	// todo close db
	db.Close()

	return nil
}
