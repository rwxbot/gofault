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
}

// CacheBackend is the storage a CacheMiddleware delegates to. It exists so the
// cache can live outside the process: *InMemoryCache serves a single instance,
// while a Redis-backed implementation lets several instances share one cache.
type CacheBackend interface {
	// Get returns a cached response. ok is false on a miss, and Get must not
	// treat a miss as an error.
	Get(key string) (value []byte, statusCode int, headers map[string]string, ok bool)
	// Set stores a response under key.
	Set(key string, value []byte, statusCode int, headers map[string]string)
	// Delete removes a key.
	Delete(key string)
}

// InMemoryCache implements a simple in-memory cache with TTL and size limits.
type InMemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	config  CacheConfig
}

// NewInMemoryCache creates a new in-memory cache.
func NewInMemoryCache(config CacheConfig) *InMemoryCache {
	if config.TTL == 0 {
		config.TTL = 5 * time.Minute
	}
	return &InMemoryCache{
		entries: make(map[string]*cacheEntry),
		config:  config,
	}
}

// Get retrieves a cached entry by key.
func (c *InMemoryCache) Get(key string) ([]byte, int, map[string]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, 0, nil, false
	}

	// Check expiration
	if time.Now().After(entry.Expires) {
		return nil, 0, nil, false
	}

	return entry.Value, entry.StatusCode, entry.Headers, true
}

// Set stores a value in the cache.
func (c *InMemoryCache) Set(key string, value []byte, statusCode int, headers map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if at capacity
	if c.config.MaxSize > 0 && len(c.entries) >= c.config.MaxSize {
		c.evictOldest()
	}

	now := time.Now()
	c.entries[key] = &cacheEntry{
		Value:      value,
		StatusCode: statusCode,
		Headers:    headers,
		Created:    now,
		Expires:    now.Add(c.config.TTL),
	}
}

// Delete removes an entry from the cache.
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries from the cache.
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*cacheEntry)
}

// Size returns the number of entries in the cache.
func (c *InMemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// evictOldest removes the oldest entry from the cache.
func (c *InMemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range c.entries {
		if first || entry.Created.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Created
			first = false
		}
	}

	if oldestKey != "" {
		delete(c.entries, oldestKey)
	}
}

// CacheMiddleware creates a caching middleware.
func CacheMiddleware(cache CacheBackend, config CacheConfig) core.MiddlewareFunc {
	if config.SkipFunc == nil {
		config.SkipFunc = DefaultCacheConfig().SkipFunc
	}
	maxBody := config.MaxBodyBytes
	if maxBody <= 0 {
		maxBody = DefaultMaxBodyBytes
	}

	return func(ctx *core.Ctx, next core.Handler) error {
		// Check if we should skip caching
		if config.SkipFunc(ctx) {
			return next(ctx)
		}

		// Generate cache key
		cacheKey := generateCacheKey(ctx)

		// Try to get from cache
		if body, statusCode, headers, ok := cache.Get(cacheKey); ok {
			// Serve from cache
			ctx.Response.Header().Set("X-Cache", "HIT")
			for k, v := range headers {
				ctx.Response.Header().Set(k, v)
			}
			ctx.Response.WriteHeader(statusCode)
			ctx.Response.Write(body)
			return nil
		}

		// Capture response
		rec := &responseCapture{
			ResponseWriter: ctx.Response,
			statusCode:     200,
			body:           []byte{},
			headers:        make(http.Header),
			maxBody:        maxBody,
		}
		ctx.Response = rec
		// Restore the real writer before returning, so anything writing to the
		// response after this middleware returns (the router's exception filter,
		// for one) is not captured into this handler instead of reaching the
		// client.
		defer func() { ctx.Response = rec.ResponseWriter }()

		// Call next handler
		err := next(ctx)

		// The capture buffers everything: Header() hands out the private map,
		// so the handler's headers only reach the real writer if we copy them
		// before the status line is committed. Delegating WriteHeader/Write
		// directly used to commit the real response while the handler's
		// headers were still sitting in the private map — every header a
		// handler set (Content-Type included) was dropped on a MISS.
		if err != nil {
			// Discard the buffered output and let the router's error path
			// (exception filter or 500) answer; replaying a half-written
			// 2xx would commit a status line the filter cannot retract.
			return err
		}

		// A response that outgrew the buffer was already streamed to the
		// client; it must not be replayed or cached.
		if rec.streamed {
			return nil
		}

		for k, values := range rec.headers {
			for _, v := range values {
				rec.ResponseWriter.Header().Add(k, v)
			}
		}
		// Only cacheable responses carry the MISS marker, matching the old
		// semantics: an error response is replayed untouched and never
		// enters the cache.
		cacheable := rec.statusCode < 400 && !uncacheableResponse(rec.headers)
		if cacheable {
			rec.ResponseWriter.Header().Set("X-Cache", "MISS")
		}
		rec.ResponseWriter.WriteHeader(rec.statusCode)
		if len(rec.body) > 0 {
			_, _ = rec.ResponseWriter.Write(rec.body)
		}

		// Cache the response if successful
		if cacheable {
			headers := make(map[string]string)
			for k, v := range rec.headers {
				if len(v) > 0 {
					headers[k] = v[0]
				}
			}
			cache.Set(cacheKey, rec.body, rec.statusCode, headers)
		}

		return err
	}
}
