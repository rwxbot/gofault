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

func TestDefaultKeyFunc_IgnoresForgedProxyHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "203.0.113.7:54321"
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	req.Header.Set("X-Real-IP", "10.0.0.1")
	ctx := &core.Ctx{Request: req}

	// With no trusted proxies the TCP address is authoritative: a client
	// rotating forged headers must not get a fresh bucket per request.
	key := DefaultKeyFunc(ctx)
	if key != "203.0.113.7:54321" {
		t.Errorf("key = %s, want 203.0.113.7:54321", key)
	}
}

func TestRateLimiter_ForgedXFFCannotBypass(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{RequestsPerSecond: 1, BurstSize: 2})
	mw := rl.Middleware()

	// Same connection, rotating XFF values: all requests share one bucket,
	// so the third must be limited despite the header churn.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "203.0.113.7:54321"
		req.Header.Set("X-Forwarded-For", "10.0.0."+string(rune('1'+i)))
		ctx := &core.Ctx{Request: req}
		if err := mw(ctx, func(ctx *core.Ctx) error { return nil }); err != nil {
			t.Fatalf("request %d unexpectedly limited: %v", i+1, err)
		}
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "203.0.113.7:54321"
	req.Header.Set("X-Forwarded-For", "172.16.9.9")
	ctx := &core.Ctx{Request: req}
	if err := mw(ctx, func(ctx *core.Ctx) error { return nil }); err == nil {
		t.Error("third request from same IP should be limited despite forged XFF rotation")
	}
}

func TestRateLimiter_TrustedProxiesUseXFFChain(t *testing.T) {
	rl := NewRateLimiter(RateLimiterConfig{
		RequestsPerSecond: 1,
		BurstSize:         1,
		TrustedProxies:    []string{"10.0.0.0/8"},
	})
	mw := rl.Middleware()

	// Connection comes from a trusted proxy, so the rightmost untrusted XFF
	// hop is the client. Rotating the LEFT side must not change the key.
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.1.2.3:1234"
		req.Header.Set("X-Forwarded-For", "192.168.5."+string(rune('1'+i))+", 203.0.113.9")
		ctx := &core.Ctx{Request: req}
		if i == 0 {
			if err := mw(ctx, func(ctx *core.Ctx) error { return nil }); err != nil {
				t.Fatalf("first request limited: %v", err)
			}
			continue
		}
		if err := mw(ctx, func(ctx *core.Ctx) error { return nil }); err == nil {
			t.Error("second request from same real client (203.0.113.9) should be limited")
		}
	}
}
