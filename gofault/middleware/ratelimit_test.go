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