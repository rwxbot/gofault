package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultOpenAPIConfig(t *testing.T) {
	cfg := DefaultOpenAPIConfig()

	if !cfg.Enabled {
		t.Error("Enabled should be true")
	}
	if cfg.Path != "/openapi.json" {
		t.Errorf("Path = %s, want /openapi.json", cfg.Path)
	}
	if cfg.Title != "API" {
		t.Errorf("Title = %s, want API", cfg.Title)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("Version = %s, want 1.0.0", cfg.Version)
	}
}

func TestOpenAPIMiddleware_Disabled(t *testing.T) {
	cfg := DefaultOpenAPIConfig()
	cfg.Enabled = false

	if handler := OpenAPIMiddleware(cfg); handler != nil {
		t.Error("disabled middleware should return a nil handler")
	}
}

func TestOpenAPIMiddleware_ServesSpec(t *testing.T) {
	cfg := DefaultOpenAPIConfig()

	handler := OpenAPIMiddleware(cfg)

	req := httptest.NewRequest("GET", "/openapi.json", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	if err := handler(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %s, want application/json", ct)
	}

	var spec OpenAPIDocument
	if err := json.Unmarshal(rec.Body.Bytes(), &spec); err != nil {