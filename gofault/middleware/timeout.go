package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofault/gofault/core"
)

// TimeoutConfig holds configuration for the timeout middleware.
type TimeoutConfig struct {
	// Enabled enables the timeout middleware.
	Enabled bool
	// Duration is the timeout duration.
	Duration time.Duration
	// ErrorMessage is the message to return on timeout.
	ErrorMessage string
}

// DefaultTimeoutConfig returns a default timeout configuration.