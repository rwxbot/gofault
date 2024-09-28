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