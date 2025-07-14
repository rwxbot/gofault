package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestCORS_DefaultConfig(t *testing.T) {
	cfg := DefaultCORSConfig()
	if len(cfg.AllowOrigins) != 1 || cfg.AllowOrigins[0] != "*" {
		t.Errorf("default AllowOrigins should be [*], got %v", cfg.AllowOrigins)
	}
	if len(cfg.AllowMethods) != 6 {
		t.Errorf("default AllowMethods should have 6 methods, got %d", len(cfg.AllowMethods))
	}
	if cfg.MaxAge != 86400 {
		t.Errorf("default MaxAge should be 86400, got %d", cfg.MaxAge)
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	cfg := DefaultCORSConfig()
	middleware := CORS(cfg)

	next := func(ctx *core.Ctx) error {
		t.Error("next should not be called for OPTIONS preflight")
		return nil
	}

	req := httptest.NewRequest("OPTIONS", "/api/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := middleware(ctx, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if ctx.StatusCode != 204 {
		t.Errorf("StatusCode = %d, want 204", ctx.StatusCode)
	}

	// Check headers
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %s, want *", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("Access-Control-Allow-Methods should be set")
	}
}

func TestCORS_AllowOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins: []string{"http://example.com"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type"},
	}
	middleware := CORS(cfg)

	next := func(ctx *core.Ctx) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Errorf("Access-Control-Allow-Origin = %s, want http://example.com", got)
	}
}

func TestCORS_WildcardOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins: []string{"*"},
	}
	middleware := CORS(cfg)

	next := func(ctx *core.Ctx) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %s, want *", got)
	}
}

func TestCORS_SubdomainWildcard(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins: []string{"*.example.com"},
	}
	middleware := CORS(cfg)

	next := func(ctx *core.Ctx) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://api.example.com")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://api.example.com" {
		t.Errorf("Access-Control-Allow-Origin = %s, want http://api.example.com", got)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins: []string{"http://allowed.com"},
	}
	middleware := CORS(cfg)

	next := func(ctx *core.Ctx) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://disallowed.com")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %s, want empty", got)
	}
}

func TestCORS_Credentials(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins:     []string{"http://example.com"},
		AllowCredentials: true,
	}
	middleware := CORS(cfg)

	next := func(ctx *core.Ctx) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %s, want true", got)
	}
}

func TestCORS_ExposeHeaders(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins:  []string{"*"},
		ExposeHeaders: []string{"X-Custom-Header", "X-Request-ID"},
	}
	middleware := CORS(cfg)

	next := func(ctx *core.Ctx) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	expected := "X-Custom-Header, X-Request-ID"
	if got := w.Header().Get("Access-Control-Expose-Headers"); got != expected {
		t.Errorf("Access-Control-Expose-Headers = %s, want %s", got, expected)
	}
}

func TestCORS_ForwardToNext(t *testing.T) {
	cfg := DefaultCORSConfig()
	middleware := CORS(cfg)

	nextCalled := false
	next := func(ctx *core.Ctx) error {
		nextCalled = true
		return nil
	}

	req := httptest.NewRequest("POST", "/api/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)
	ctx.StatusCode = http.StatusCreated

	err := middleware(ctx, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Error("next handler should be called")
	}
	if ctx.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", ctx.StatusCode, http.StatusCreated)
	}
}
