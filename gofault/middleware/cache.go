package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gofault/gofault/core"
)

// CacheConfig holds configuration for the cache middleware.
type CacheConfig struct {
	// TTL is the time-to-live for cache entries.
	TTL time.Duration
	// MaxSize is the maximum number of entries in the cache (0 = unlimited).
	MaxSize int
	// SkipFunc is an optional function to determine if a request should be skipped.
	SkipFunc func(*core.Ctx) bool
}

// DefaultCacheConfig returns a default cache configuration.
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL:     5 * time.Minute,
		MaxSize: 1000,
		SkipFunc: func(ctx *core.Ctx) bool {
			// Skip non-GET requests by default
			return ctx.Request.Method != "GET"
		},
	}
}

// cacheEntry represents a cached response.
type cacheEntry struct {
	Value      []byte
	StatusCode int
	Headers    map[string]string
	Created    time.Time
	Expires    time.Time
}

// InMemoryCache implements a simple in-memory cache with TTL and size limits.
type InMemoryCache struct {
	mu      sync.RWMutex