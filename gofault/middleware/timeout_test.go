package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofault/gofault/core"
)

func TestDefaultTimeoutConfig(t *testing.T) {
	config := DefaultTimeoutConfig()
	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.Duration != 30*time.Second {
		t.Errorf("Expected Duration 30s, got %v", config.Duration)
	}
}

func TestTimeoutMiddleware_Disabled(t *testing.T) {
	config := DefaultTimeoutConfig()
	config.Enabled = false

	handler := TimeoutMiddleware(config)
	if handler != nil {
		t.Error("Expected nil handler when disabled")
	}
}

func TestTimeoutMiddleware_NoTimeout(t *testing.T) {
	middleware := TimeoutMiddleware(DefaultTimeoutConfig())

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

func TestTimeoutMiddleware_Timeout(t *testing.T) {
	config := DefaultTimeoutConfig()
	config.Duration = 50 * time.Millisecond

	middleware := TimeoutMiddleware(config)

	next := func(ctx *core.Ctx) error {
		time.Sleep(200 * time.Millisecond)
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("Expected status %d, got %d", http.StatusGatewayTimeout, w.Code)
	}
}
