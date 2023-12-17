package middlewares

import (
	"net/http"
	"time"

	"github.com/SerjRamone/gophermart/pkg/logger"
	"go.uber.org/zap"
)

type (
	// stores response data
	responseData struct {
		status int
		size   int
	}

	// http.ResponseWriter implementation
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write ...
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// write response via original http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // get size
	return size, err
}

// WriteHeader ...
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// write status code via original http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // get status code
}

// RequestLogger middleware for request logging
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		next.ServeHTTP(&lw, r)

		// get request duration
		duration := time.Since(start)

		logger.Info("handle request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Int("status", responseData.status),
			zap.Duration("duration", duration),
			zap.Int("size", responseData.size),
		)
	})
}
