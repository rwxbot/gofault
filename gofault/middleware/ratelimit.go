package middleware

import (
	"net"
	"sync"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

// Limiter decides whether a given key may pass. It exists so the middleware is
// not tied to RateLimiter's in-process buckets, which do not work when the
// application runs as more than one replica: each process counts separately, so
// the effective limit is N times the configured one.
//
// Use RateLimiter for a single process, or a Redis-backed limiter when deployed
// as more than one.
type Limiter interface {
	// Allow reports whether key may proceed.
	Allow(key string) bool
	// Close releases any resources the limiter holds.
	Close()
}

// RateLimiterConfig holds rate limiter configuration.
type RateLimiterConfig struct {
	// RequestsPerSecond is the number of requests allowed per second per key.
	RequestsPerSecond float64
	// BurstSize is the maximum burst size (number of requests that can be made at once).
	BurstSize int
	// KeyFunc extracts the rate limit key from the request (e.g., IP address).
	KeyFunc func(*core.Ctx) string
}
