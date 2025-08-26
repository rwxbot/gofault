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

// DefaultKeyFunc returns the client IP as the rate limit key.
func DefaultKeyFunc(ctx *core.Ctx) string {
	ip := ctx.Request.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = ctx.Request.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = ctx.Request.RemoteAddr
	}
	return ip
}

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	config  RateLimiterConfig
	buckets map[string]*tokenBucket
	mu      sync.RWMutex
}

type tokenBucket struct {
	tokens     float64
	lastUpdate time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(cfg RateLimiterConfig) *RateLimiter {
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = DefaultKeyFunc
	}
	if cfg.BurstSize == 0 {
		cfg.BurstSize = 10
	}
	if cfg.RequestsPerSecond == 0 {
		cfg.RequestsPerSecond = 100
	}

	rl := &RateLimiter{
		config:  cfg,
		buckets: make(map[string]*tokenBucket),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}