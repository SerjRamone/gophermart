// Package logger ...
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is no-operation logger by default
// Init in Init() function
var log *zap.Logger = zap.NewNop()

// Init and configure logger
func Init(level string) error {
	// set log level
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	log = zl
	return nil
}

// Info logs a message at InfoLevel
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Debug logs a message at DebugLevel
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

// Warn logs a message at WarnLevel
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Fatal logs a message at FatalLevel, calls os.Exit(1)
func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}
