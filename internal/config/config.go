// Package config ...
package config

import (
	"flag"

	"github.com/caarlos0/env"
	"go.uber.org/zap/zapcore"
)

const (
	defaultRunAddress           = "localhost:8080"
	defaultDatabaseURI          = ""
	defaultAccrualSystemAddress = "localhost:8088"
	defaultLogLevel             = "error"
	defaultSecretKey            = ""
	defaultTokenExpiration      = 3600

	usageRunAddress           = "address and port for running app"
	usageDatabaseURI          = "database URI"
	usageAccrualSystemAddress = "address and port of accrual system"
	usageLogLevel             = "log level (`error` by default)"
	usageSecretKey            = "secret key for encoders"
	usageTokenExpiration      = "authorization token expiration time (3600 sec by default)"
)

// Gophermart is a gophermart app config
type Gophermart struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel             string `env:"LOG_LEVEL"`
	SecretKey            string `env:"SECRET_KEY"`
	TokenExpiration      int    `env:"TOKEN_EXPIRATION"`
}

// NewGophermart constructor for gophermart config
func NewGophermart() (Gophermart, error) {
	g := Gophermart{}
	g.parseFlags()
	return g, g.parseEnv()
}

// parseFlags parse cli flags
func (g *Gophermart) parseFlags() {
	flag.StringVar(&g.RunAddress, "a", defaultRunAddress, usageRunAddress)
	flag.StringVar(&g.DatabaseURI, "d", defaultDatabaseURI, usageDatabaseURI)
	flag.StringVar(&g.AccrualSystemAddress, "r", defaultAccrualSystemAddress, usageAccrualSystemAddress)
	flag.StringVar(&g.LogLevel, "l", defaultLogLevel, usageLogLevel)
	flag.StringVar(&g.SecretKey, "s", defaultSecretKey, usageSecretKey)
	flag.IntVar(&g.TokenExpiration, "e", defaultTokenExpiration, usageTokenExpiration)

	flag.Parse()
}

// parseEnv parse environtment variables
func (g *Gophermart) parseEnv() error {
	return env.Parse(g)
}

// MarshalLogObject zapcore.ObjectMarshaler implemet for loggin agent config struct
func (g *Gophermart) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("RunAddress", g.RunAddress)
	enc.AddString("DatabaseURI", g.DatabaseURI)
	enc.AddString("AccrualSystemAddress", g.AccrualSystemAddress)
	enc.AddString("LogLevel", g.LogLevel)
	enc.AddString("SecretKey", g.SecretKey)
	enc.AddInt("TokenExpiration", g.TokenExpiration)

	return nil
}
