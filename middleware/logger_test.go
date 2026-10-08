package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultRequestLoggerConfig(t *testing.T) {
	config := DefaultRequestLoggerConfig()
	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.LogBody {
		t.Error("Expected LogBody to be false")
	}
	if config.LogHeaders {
		t.Error("Expected LogHeaders to be false")
	}
}

func TestRequestLoggerMiddleware_Disabled(t *testing.T) {
	config := DefaultRequestLoggerConfig()
	config.Enabled = false

	handler := RequestLoggerMiddleware(config)
	if handler != nil {
		t.Error("Expected nil handler when disabled")
	}
}

func TestRequestLoggerMiddleware_LogRequest(t *testing.T) {
	middleware := RequestLoggerMiddleware(DefaultRequestLoggerConfig())

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte(`{"status":"ok"}`))
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test?foo=bar", nil)
	req.Header.Set("X-Custom-Header", "test-value")
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

func TestRequestLoggerMiddleware_WithBodyLogging(t *testing.T) {
	config := DefaultRequestLoggerConfig()
	config.LogBody = true

	middleware := RequestLoggerMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusCreated)
		return nil
	}

	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := middleware(ctx, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRequestLoggerMiddleware_WithHeaderLogging(t *testing.T) {
	config := DefaultRequestLoggerConfig()
	config.LogHeaders = true

	middleware := RequestLoggerMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := middleware(ctx, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
