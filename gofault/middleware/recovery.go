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