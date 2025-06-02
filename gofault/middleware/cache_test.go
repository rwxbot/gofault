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