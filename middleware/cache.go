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

// uncacheableResponse reports whether the handler's own headers forbid
// caching. A handler that sets Cache-Control: no-store, no-cache or private
// — a per-user dashboard, a session endpoint — used to be cached anyway:
// the cacheability check looked only at the status code, and the next
// request, authenticated or not, was served the first caller's body.
// A response carrying Vary is also skipped: the key cannot know which
// request headers the response varies on.
func uncacheableResponse(h http.Header) bool {
	if len(h.Values("Vary")) > 0 {
		return true
	}
	for _, cc := range h.Values("Cache-Control") {
		for _, directive := range strings.Split(cc, ",") {
			switch strings.ToLower(strings.TrimSpace(directive)) {
			case "no-store", "no-cache", "private", "max-age=0":
				return true
			}
		}
	}
	return false
}

// generateCacheKey generates a cache key from the request.
//
// Host is part of the key: without it, a deployment behind a reverse proxy
// or any multi-host setup collides distinct tenants onto one entry — a
// request to a.example.com/foo could be served the body cached for
// b.example.com/foo.
//
// The request's credentials (Authorization, Cookie) are hashed into the key:
// a key built only from Host+Method+Path+Query serves the first caller's
// authenticated body to every later caller hitting the same URL, including
// unauthenticated ones. Hashing keeps raw secrets out of the key — keys are
// stored verbatim, including in a shared Redis backend.
func generateCacheKey(ctx *core.Ctx) string {
	key := ctx.Request.Host + "|" + ctx.Request.Method + ":" + ctx.Request.URL.Path + ":" + ctx.Request.URL.RawQuery

	auth := ctx.Request.Header.Get("Authorization")
	cookie := ctx.Request.Header.Get("Cookie")
	if auth != "" || cookie != "" {
		sum := sha256.Sum256([]byte(auth + "\x00" + cookie))
		key += "|cred:" + hex.EncodeToString(sum[:])
	}
	return key
}

// responseCapture captures the response for caching.
//
// responseCapture buffers a handler's response so it can be cached and then
// replayed onto the real writer. Nothing touches the real ResponseWriter
// until the middleware replays it: delegating WriteHeader/Write directly
// while Header() returned a private map committed the status line before the
// handler's headers existed, so those headers were silently dropped.
//
// Buffering is capped at maxBody bytes. A response that outgrows the cap —
// or that Flushes, or hijacks the connection — is streamed to the client
// instead and excluded from the cache: embedding http.ResponseWriter does
// NOT make the capture satisfy http.Flusher or http.Hijacker (the method
// set comes from the static embedded type), so handlers that type-assert
// those interfaces used to fail outright under the cache middleware.
type responseCapture struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	body        []byte
	headers     http.Header
	maxBody     int
	streamed    bool // bytes already reached the client; the response will not be cached
}

// streamOut commits the buffered response onto the real writer.
func (r *responseCapture) streamOut() {
	for k, values := range r.headers {
		for _, v := range values {
			r.ResponseWriter.Header().Add(k, v)
		}
	}
	if !r.wroteHeader {
		r.statusCode = http.StatusOK
		r.wroteHeader = true
	}
	r.ResponseWriter.WriteHeader(r.statusCode)
	if len(r.body) > 0 {
		_, _ = r.ResponseWriter.Write(r.body)
		r.body = nil
	}
}

func (r *responseCapture) WriteHeader(statusCode int) {
	if !r.wroteHeader {
		r.statusCode = statusCode
		r.wroteHeader = true
	}
}

func (r *responseCapture) Write(b []byte) (int, error) {
	if r.streamed {
		// Past the cap (or past a Flush/Hijack): the response is being
		// streamed, pass writes straight through.
		return r.ResponseWriter.Write(b)
	}
	if !r.wroteHeader {
		r.statusCode = http.StatusOK
		r.wroteHeader = true
	}
	if len(r.body)+len(b) > r.maxBody {
		r.streamOut()
		r.streamed = true
		return r.ResponseWriter.Write(b)
	}
	r.body = append(r.body, b...)
	return len(b), nil
}

func (r *responseCapture) Header() http.Header {
	if r.headers == nil {
		r.headers = make(http.Header)
	}
	return r.headers
}

// Flush makes the capture satisfy http.Flusher explicitly: a handler that
// type-asserts ctx.Response to http.Flusher used to fail under the cache
// middleware. Flushing commits the buffered response and opts the request
// out of caching — a streaming handler's response is on the wire already.
func (r *responseCapture) Flush() {
	if !r.streamed {
		r.streamOut()
		r.streamed = true
	}
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack makes the capture satisfy http.Hijacker, delegating to the real
// writer when it supports it. A hijacked connection is never cached.
func (r *responseCapture) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("middleware/cache: underlying ResponseWriter does not implement http.Hijacker")
	}
	r.streamed = true
	return h.Hijack()
}
