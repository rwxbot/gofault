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

// A panicking handler must not kill the process: the goroutine's panic is
// converted to an error and the buffered response is flushed normally.
func TestTimeoutMiddleware_HandlerPanicRecovered(t *testing.T) {
	config := DefaultTimeoutConfig()
	config.Duration = time.Second

	middleware := TimeoutMiddleware(config)

	next := func(ctx *core.Ctx) error {
		panic("handler exploded")
	}

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	done := make(chan error, 1)
	go func() { done <- middleware(ctx, next) }()

	select {
	case err := <-done:
		if err == nil {
			t.Error("expected the panic to surface as an error")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("middleware did not return after the handler panicked")
	}
}

// The timeout branch writes the 504 itself, so it must return nil: returning
// the context error makes the router write a second response onto the same
// writer.
func TestTimeoutMiddleware_TimeoutReturnsNil(t *testing.T) {
	config := DefaultTimeoutConfig()
	config.Duration = 50 * time.Millisecond

	middleware := TimeoutMiddleware(config)

	next := func(ctx *core.Ctx) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	if err := middleware(ctx, next); err != nil {
		t.Errorf("timeout branch should return nil (504 already written), got %v", err)
	}
	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("Expected status %d, got %d", http.StatusGatewayTimeout, w.Code)
	}
	if ctx.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("ctx.StatusCode = %d, want %d for the logger/metrics middleware", ctx.StatusCode, http.StatusGatewayTimeout)
	}
}
