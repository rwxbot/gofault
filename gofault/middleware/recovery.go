package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gofault/gofault/core"
)

// RecoveryConfig holds configuration for the recovery middleware.
type RecoveryConfig struct {
	// Enabled enables the recovery middleware.
	Enabled bool
	// StackTrace enables stack trace in error response (for debugging).
	StackTrace bool
}

// DefaultRecoveryConfig returns a default recovery configuration.
func DefaultRecoveryConfig() RecoveryConfig {
	return RecoveryConfig{
		Enabled:    true,
		StackTrace: false,
	}
}

// RecoveryMiddleware recovers from panics and returns a proper error response.
func RecoveryMiddleware(config RecoveryConfig) core.MiddlewareFunc {
	if !config.Enabled {
		return nil
	}

	return func(ctx *core.Ctx, next core.Handler) error {
		defer func() {
			if r := recover(); r != nil {
				var errMsg string
				switch val := r.(type) {
				case error:
					errMsg = val.Error()
				case string:
					errMsg = val
				default:
					errMsg = fmt.Sprintf("panic: %v", val)
				}

				stackTrace := ""
				if config.StackTrace {
					stackTrace = string(debug.Stack())
				}

				ctx.Response.Header().Set("Content-Type", "application/json")
				ctx.Response.WriteHeader(http.StatusInternalServerError)

				response := map[string]any{
					"status":  "error",
					"message": errMsg,
				}
				if stackTrace != "" {
					response["stack"] = stackTrace
				}

				json.NewEncoder(ctx.Response).Encode(response)
			}
		}()

		return next(ctx)
	}
}
