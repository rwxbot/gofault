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