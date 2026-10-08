package middleware

import (
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

		if config.LogBody {
			logData["body"] = readRequestBody(ctx)
		}

		// Call next handler
		err := next(ctx)

		// Log response
		duration := time.Since(start)
		logData["duration_ms"] = duration.Milliseconds()
		logData["status"] = ctx.StatusCode

		if config.LogBody {
			logData["response_body"] = "response captured"
		}

		// In a real implementation, this would use the app's logger
		// For now, we just avoid blocking

		return err
	}
}

// formatHeaders formats HTTP headers as a map.
func formatHeaders(headers map[string][]string) map[string]string {
	result := make(map[string]string)
	for k, v := range headers {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// readRequestBody reads the request body (simplified - in production use ctx.Request.Body)
func readRequestBody(ctx *core.Ctx) string {
	return ""
}
