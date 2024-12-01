// Package middleware provides common HTTP middleware for gofault.
package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"time"

	"github.com/gofault/gofault/core"
)

// RequestLoggerConfig holds configuration for the request logger middleware.
type RequestLoggerConfig struct {
	// Enabled enables the request logger.
	Enabled bool
	// LogBody enables logging of request and response bodies.
	LogBody bool
	// LogHeaders enables logging of request headers.
	LogHeaders bool
}

// DefaultRequestLoggerConfig returns a default request logger configuration.
func DefaultRequestLoggerConfig() RequestLoggerConfig {
	return RequestLoggerConfig{
		Enabled:    true,
		LogBody:    false,
		LogHeaders: false,
	}
}

// RequestLoggerMiddleware creates a middleware that logs HTTP requests and responses.
func RequestLoggerMiddleware(config RequestLoggerConfig) core.MiddlewareFunc {
	if !config.Enabled {
		return nil
	}

	return func(ctx *core.Ctx, next core.Handler) error {
		start := time.Now()

		// Log request
		logData := map[string]any{
			"method":  ctx.Request.Method,
			"path":    ctx.Request.URL.Path,
			"query":   ctx.Request.URL.RawQuery,
			"ip":      ctx.Request.RemoteAddr,
			"started": start.Format(time.RFC3339),
		}

		if config.LogHeaders {
			logData["headers"] = formatHeaders(ctx.Request.Header)
		}
