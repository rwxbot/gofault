package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultRecoveryConfig(t *testing.T) {
	config := DefaultRecoveryConfig()
	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.StackTrace {
		t.Error("Expected StackTrace to be false")
	}
}

func TestRecoveryMiddleware_PanicRecovered(t *testing.T) {
	middleware := RecoveryMiddleware(DefaultRecoveryConfig())

	next := func(ctx *core.Ctx) error {
		panic("test panic")
	}

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "test panic") {
		t.Errorf("Expected body to contain 'test panic', got: %s", body)
	}
}

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	middleware := RecoveryMiddleware(DefaultRecoveryConfig())

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte(`{"status":"ok"}`))
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
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

func TestRecoveryMiddleware_StackTrace(t *testing.T) {
	config := DefaultRecoveryConfig()
	config.StackTrace = true

	middleware := RecoveryMiddleware(config)

	next := func(ctx *core.Ctx) error {
		panic("test panic with stack")
	}

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "goroutine") {
		t.Errorf("Expected body to contain stack trace, got: %s", body)
	}
}

func TestRecoveryMiddleware_Disabled(t *testing.T) {
	config := DefaultRecoveryConfig()
	config.Enabled = false

	handler := RecoveryMiddleware(config)
	if handler != nil {
		t.Error("Expected nil handler when disabled")
	}
}

func TestRecoveryMiddleware_PanicWithError(t *testing.T) {
	middleware := RecoveryMiddleware(DefaultRecoveryConfig())

	next := func(ctx *core.Ctx) error {
		panic(http.ErrAbortHandler)
	}

	req := httptest.NewRequest(http.MethodGet, "/panic-error", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
