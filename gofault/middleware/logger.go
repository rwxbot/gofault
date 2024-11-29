package middleware

import (
	"time"

	"github.com/gofault/gofault/core"
)

// RequestLoggerConfig holds configuration for the request logger middleware.
type RequestLoggerConfig struct {
	// Enabled enables the request logger.