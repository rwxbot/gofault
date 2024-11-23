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