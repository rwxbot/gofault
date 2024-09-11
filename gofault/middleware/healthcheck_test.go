package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultHealthCheckConfig(t *testing.T) {
	cfg := DefaultHealthCheckConfig()

	if cfg.Path != "/health" {
		t.Errorf("Path = %s, want /health", cfg.Path)
	}
	if !cfg.IncludeTimestamp {
		t.Error("IncludeTimestamp should be true")
	}
	if !cfg.IncludeChecks {
		t.Error("IncludeChecks should be true")
	}
}

func TestHealthCheckMiddleware_Up(t *testing.T) {
	cfg := DefaultHealthCheckConfig()
	cfg.Checks = map[string]HealthChecker{
		"db": func() HealthCheck {
			return HealthCheck{Status: HealthStatusUp}
		},
	}

	handler := HealthCheckMiddleware(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	ctx := &core.Ctx{Request: req, Response: rec}

	handler(ctx)

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	var response HealthCheckResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}