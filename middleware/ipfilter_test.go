package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultIPFilterConfig(t *testing.T) {
	config := DefaultIPFilterConfig()
	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.Mode != "block" {
		t.Errorf("Expected Mode 'block', got '%s'", config.Mode)
	}
}

func TestIPFilterMiddleware_Disabled(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Enabled = false

	handler := IPFilterMiddleware(config)
	if handler != nil {
		t.Error("Expected nil handler when disabled")
	}
}

func TestIPFilterMiddleware_NoRestrictions(t *testing.T) {
	middleware := IPFilterMiddleware(DefaultIPFilterConfig())

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte(`{"status":"ok"}`))
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := middleware(ctx, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestIPFilterMiddleware_BlockList(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Block = []string{"192.168.1.100"}

	middleware := IPFilterMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)
	ctx.Request.RemoteAddr = "192.168.1.100:12345"

	middleware(ctx, next)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestIPFilterMiddleware_AllowList(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Allow = []string{"10.0.0.0/8"}

	middleware := IPFilterMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte(`{"status":"ok"}`))
		return nil
	}

	// IP in allow range
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)
	ctx.Request.RemoteAddr = "10.1.2.3:12345"

	err := middleware(ctx, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// IP not in allow range
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	ctx2 := core.NewCtx(w2, req2)
	ctx2.Request.RemoteAddr = "192.168.1.1:12345"

	middleware(ctx2, func(c *core.Ctx) error {
		c.Response.WriteHeader(http.StatusOK)
		return nil
	})

	if w2.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w2.Code)
	}
}

func TestParseIPPatterns(t *testing.T) {
	patterns := []string{"192.168.1.1", "10.0.0.0/8", "172.16.0.0/12"}
	nets := parseIPPatterns(patterns)

	if len(nets) != 3 {
		t.Errorf("Expected 3 networks, got %d", len(nets))
	}
}
