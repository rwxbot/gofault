package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/prometheus/client_golang/prometheus"
)

func TestDefaultMetricsConfig(t *testing.T) {
	cfg := DefaultMetricsConfig()
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.Namespace != "gofault" {
		t.Errorf("expected namespace 'gofault', got '%s'", cfg.Namespace)
	}
	if cfg.Subsystem != "http" {
		t.Errorf("expected subsystem 'http', got '%s'", cfg.Subsystem)
	}
	if !cfg.SkipHealthCheck {
		t.Error("expected SkipHealthCheck to be true")
	}
	if len(cfg.Buckets) == 0 {
		t.Error("expected non-empty Buckets")
	}
}

func TestMetricsMiddleware_Disabled(t *testing.T) {
	cfg := MetricsConfig{Enabled: false}
	mw, reg := MetricsMiddleware(cfg)
	if mw != nil {
		t.Error("expected nil middleware when disabled")
	}
	if reg != nil {
		t.Error("expected nil registry when disabled")
	}
}

func TestMetricsMiddleware_RecordsRequest(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.SkipHealthCheck = false // record all paths

	mw, registry := MetricsMiddleware(cfg)
	if mw == nil {
		t.Fatal("middleware is nil")
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	var called bool