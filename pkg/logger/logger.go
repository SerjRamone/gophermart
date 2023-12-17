// Package logger ...
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is no-operation logger by default
// Init in New() function
var log *zap.Logger = zap.NewNop()

// Option ...
type Option func(*zap.Config) error

// New init and configure logger
func New(opts ...Option) error {
	// set default config
	cfg := zap.NewProductionConfig()

	// apply options
	for _, fn := range opts {
		if err := fn(&cfg); err != nil {
			return err
		}
	}

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	log = zl
	return nil
}

// LevelOption return Option func for setting Loglevel value
func LevelOption(level string) Option {
	return func(c *zap.Config) error {
		lvl, err := zap.ParseAtomicLevel(level)
		if err != nil {
			return err
		}
		c.Level = lvl
		return nil
	}
}

// TimeKeyOption return Option func for setting TimeKey value
func TimeKeyOption(timeKey string) Option {
	return func(c *zap.Config) error {
		c.EncoderConfig.TimeKey = timeKey
		return nil
	}
}

// TimeEncoderOption return Option func for setting TimeEncoder value
func TimeEncoderOption(encoder zapcore.TimeEncoder) Option {
	return func(c *zap.Config) error {
		c.EncoderConfig.EncodeTime = encoder
		return nil
	}
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
