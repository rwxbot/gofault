package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestRateLimiter_Allow(t *testing.T) {
	cfg := RateLimiterConfig{
		RequestsPerSecond: 10,
		BurstSize:         5,
		KeyFunc:           func(ctx *core.Ctx) string { return "test-key" },
	}
	rl := NewRateLimiter(cfg)

	// First 5 requests should be allowed (burst)
	for i := 0; i < 5; i++ {
		if !rl.Allow("test-key") {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	// 6th request should be denied
	if rl.Allow("test-key") {
		t.Error("6th request should be denied")
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	cfg := RateLimiterConfig{
		RequestsPerSecond: 100,
		BurstSize:         2,
		KeyFunc:           func(ctx *core.Ctx) string { return "test-key" },
	}
	rl := NewRateLimiter(cfg)

	// Use up burst
	rl.Allow("test-key")
	rl.Allow("test-key")

	// Should be denied
	if rl.Allow("test-key") {
		t.Error("should be denied after burst")
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	cfg := RateLimiterConfig{
		RequestsPerSecond: 10,
		BurstSize:         2,
		KeyFunc:           func(ctx *core.Ctx) string { return ctx.Request.RemoteAddr },
	}
	rl := NewRateLimiter(cfg)

	// key1 uses burst
	rl.Allow("key1")
	rl.Allow("key1")

	// key2 should still have its own burst
	if !rl.Allow("key2") {
		t.Error("key2 should be allowed")
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	cfg := RateLimiterConfig{
		RequestsPerSecond: 10,
		BurstSize:         2,
		KeyFunc:           func(ctx *core.Ctx) string { return "test-ip" },
	}
	rl := NewRateLimiter(cfg)
	middleware := rl.Middleware()

	// First two requests OK
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := &core.Ctx{Request: req}
		err := middleware(ctx, func(ctx *core.Ctx) error { return nil })
		if err != nil {
			t.Errorf("request %d unexpected error: %v", i+1, err)
		}
	}

	// Third should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &core.Ctx{Request: req}
	err := middleware(ctx, func(ctx *core.Ctx) error { return nil })
	if err == nil {
		t.Error("expected rate limit error")
	}
}

func TestRateLimiter_DefaultConfig(t *testing.T) {
	cfg := RateLimiterConfig{}
	rl := NewRateLimiter(cfg)

	if rl.config.BurstSize != 10 {
		t.Errorf("default BurstSize = %d, want 10", rl.config.BurstSize)
	}
	if rl.config.RequestsPerSecond != 100 {
		t.Errorf("default RequestsPerSecond = %f, want 100", rl.config.RequestsPerSecond)
	}
	if rl.config.KeyFunc == nil {
		t.Error("default KeyFunc should not be nil")
	}
}

func TestDefaultKeyFunc(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	ctx := &core.Ctx{Request: req}

	key := DefaultKeyFunc(ctx)
	if key != "192.168.1.1" {
		t.Errorf("key = %s, want 192.168.1.1", key)
	}
}

func TestDefaultKeyFunc_XRealIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-IP", "10.0.0.1")
	ctx := &core.Ctx{Request: req}

	key := DefaultKeyFunc(ctx)
	if key != "10.0.0.1" {
		t.Errorf("key = %s, want 10.0.0.1", key)
	}
}
