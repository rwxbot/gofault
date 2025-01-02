package middleware

import (
	"errors"
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