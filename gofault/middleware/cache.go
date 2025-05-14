package middleware

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"strings"
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
	// MaxBodyBytes caps how much of a response body is buffered for
	// caching (0 = DefaultMaxBodyBytes). A response larger than this is
	// streamed to the client instead: buffering unbounded bodies let one
	// large download pin the whole entry in memory.
	MaxBodyBytes int
	// SkipFunc is an optional function to determine if a request should be skipped.
	SkipFunc func(*core.Ctx) bool
}

// DefaultMaxBodyBytes is the body size cap applied when MaxBodyBytes is 0.
const DefaultMaxBodyBytes = 256 * 1024

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