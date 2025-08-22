package middleware

import (
	"sync"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

// RateLimiterConfig holds rate limiter configuration.
type RateLimiterConfig struct {
	// RequestsPerSecond is the number of requests allowed per second per key.
	RequestsPerSecond float64
	// BurstSize is the maximum burst size (number of requests that can be made at once).
	BurstSize int
	// KeyFunc extracts the rate limit key from the request (e.g., IP address).
	KeyFunc func(*core.Ctx) string
}