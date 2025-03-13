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