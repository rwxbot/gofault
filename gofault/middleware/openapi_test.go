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