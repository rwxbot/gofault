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
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Enabled:      true,
		Duration:     30 * time.Second,
		ErrorMessage: "Request timeout",
	}
}

// timeoutBody is the payload returned when a handler exceeds its deadline.
type timeoutBody struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
