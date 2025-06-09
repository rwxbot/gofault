package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofault/gofault/core"
)

func TestNewInMemoryCache(t *testing.T) {
	cfg := DefaultCacheConfig()
	cache := NewInMemoryCache(cfg)

	if cache == nil {
		t.Fatal("cache should not be nil")
	}
	if cache.entries == nil {
		t.Error("entries map should be initialized")
	}
}

func TestInMemoryCache_SetAndGet(t *testing.T) {
	cache := NewInMemoryCache(DefaultCacheConfig())

	cache.Set("key1", []byte("value1"), 200, map[string]string{"Content-Type": "text/plain"})

	value, statusCode, headers, ok := cache.Get("key1")
	if !ok {
		t.Fatal("key1 should exist in cache")
	}
	if string(value) != "value1" {
		t.Errorf("value = %s, want value1", string(value))
	}
	if statusCode != 200 {
		t.Errorf("statusCode = %d, want 200", statusCode)
	}
	if headers["Content-Type"] != "text/plain" {
		t.Errorf("Content-Type = %s, want text/plain", headers["Content-Type"])
	}
}

func TestInMemoryCache_GetNotFound(t *testing.T) {
	cache := NewInMemoryCache(DefaultCacheConfig())

	_, _, _, ok := cache.Get("nonexistent")
	if ok {
		t.Error("nonexistent key should not be found")
	}
}

func TestInMemoryCache_Delete(t *testing.T) {
	cache := NewInMemoryCache(DefaultCacheConfig())

	cache.Set("key1", []byte("value1"), 200, nil)
	cache.Delete("key1")

	_, _, _, ok := cache.Get("key1")
	if ok {
		t.Error("key1 should be deleted")
	}
}

func TestInMemoryCache_Clear(t *testing.T) {
	cache := NewInMemoryCache(DefaultCacheConfig())

	cache.Set("key1", []byte("value1"), 200, nil)
	cache.Set("key2", []byte("value2"), 200, nil)
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Size = %d, want 0", cache.Size())
	}
}

func TestInMemoryCache_Expiry(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.TTL = 100 * time.Millisecond
	cache := NewInMemoryCache(cfg)

	cache.Set("key1", []byte("value1"), 200, nil)

	// Should exist immediately
	_, _, _, ok := cache.Get("key1")
	if !ok {
		t.Fatal("key1 should exist immediately after set")
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, _, _, ok = cache.Get("key1")
	if ok {
		t.Error("key1 should be expired")
	}
}

func TestInMemoryCache_MaxSize(t *testing.T) {
	cfg := DefaultCacheConfig()
	cfg.MaxSize = 3
	cache := NewInMemoryCache(cfg)

	cache.Set("key1", []byte("value1"), 200, nil)
	cache.Set("key2", []byte("value2"), 200, nil)
	cache.Set("key3", []byte("value3"), 200, nil)
	cache.Set("key4", []byte("value4"), 200, nil) // Should evict key1

	if cache.Size() != 3 {
		t.Errorf("Size = %d, want 3", cache.Size())
	}

	// key1 should be evicted
	_, _, _, ok := cache.Get("key1")
	if ok {
		t.Error("key1 should be evicted")
	}

	// key2, key3, key4 should exist
	_, _, _, ok = cache.Get("key4")
	if !ok {
		t.Error("key4 should exist")
	}
}

func TestDefaultCacheConfig(t *testing.T) {
	cfg := DefaultCacheConfig()

	if cfg.TTL != 5*time.Minute {
		t.Errorf("TTL = %v, want 5m", cfg.TTL)
	}
	if cfg.MaxSize != 1000 {
		t.Errorf("MaxSize = %d, want 1000", cfg.MaxSize)
	}
	if cfg.SkipFunc == nil {
		t.Error("SkipFunc should not be nil")
	}
	// Verify SkipFunc returns true for non-GET methods
	req := httptest.NewRequest("POST", "/test", nil)
	ctx := &core.Ctx{Request: req}
	if !cfg.SkipFunc(ctx) {
		t.Error("SkipFunc should return true for POST")
	}
}

func TestGenerateCacheKey(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?foo=bar", nil)
	ctx := &core.Ctx{Request: req}

	key := generateCacheKey(ctx)
	expected := "GET:/test:foo=bar"
	if key != expected {
		t.Errorf("cacheKey = %s, want %s", key, expected)
	}