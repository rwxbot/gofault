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

	if response.Status != HealthStatusUp {
		t.Errorf("status = %s, want %s", response.Status, HealthStatusUp)
	}
}

func TestHealthCheckMiddleware_Down(t *testing.T) {
	cfg := DefaultHealthCheckConfig()
	cfg.Checks = map[string]HealthChecker{
		"db": func() HealthCheck {
			return HealthCheck{Status: HealthStatusDown, Message: "connection refused"}
		},
	}

	handler := HealthCheckMiddleware(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	ctx := &core.Ctx{Request: req, Response: rec}

	handler(ctx)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}

func TestHealthCheckMiddleware_NoChecks(t *testing.T) {
	cfg := DefaultHealthCheckConfig()
	cfg.IncludeChecks = false

	handler := HealthCheckMiddleware(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	ctx := &core.Ctx{Request: req, Response: rec}

	handler(ctx)

	if rec.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestHealthCheckConfig_RegisterHealthCheck(t *testing.T) {
	cfg := DefaultHealthCheckConfig()

	checker := func() HealthCheck {
		return HealthCheck{Status: HealthStatusUp}
	}
	cfg.RegisterHealthCheck("cache", checker)

	if len(cfg.Checks) != 1 {
		t.Errorf("Checks count = %d, want 1", len(cfg.Checks))
	}
}

func TestLatencyHealthCheck(t *testing.T) {
	checker := LatencyHealthCheck("test", func() error {
		return nil
	})

	result := checker()

	if result.Status != HealthStatusUp {
		t.Errorf("status = %s, want %s", result.Status, HealthStatusUp)
	}
	if result.Latency == "" {
		t.Error("latency should be set")
	}
}

func TestLatencyHealthCheck_Error(t *testing.T) {
	checker := LatencyHealthCheck("test", func() error {
		return &testError{"connection refused"}
	})

	result := checker()

	if result.Status != HealthStatusDown {
		t.Errorf("status = %s, want %s", result.Status, HealthStatusDown)
	}
	if result.Message != "connection refused" {
		t.Errorf("message = %s, want connection refused", result.Message)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestNewHealthCheckResponse(t *testing.T) {
	response := NewHealthCheckResponse(HealthStatusUp, "ok")

	if response.Status != HealthStatusUp {
		t.Errorf("status = %s, want %s", response.Status, HealthStatusUp)
	}
	if response.Message != "ok" {
		t.Errorf("message = %s, want ok", response.Message)
	}
}
