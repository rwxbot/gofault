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