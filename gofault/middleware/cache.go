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