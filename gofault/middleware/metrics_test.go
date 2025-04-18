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
	mw(ctx, func(c *core.Ctx) error {
		called = true
		c.StatusCode = http.StatusOK
		return nil
	})

	if !called {
		t.Error("next handler was not called")
	}

	// Collect metrics from registry
	mfs, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}

	totalFound := false
	for _, mf := range mfs {
		if mf.GetName() == "gofault_http_requests_total" {
			totalFound = true
			metrics := mf.GetMetric()
			if len(metrics) == 0 {
				t.Error("no request metrics recorded")
			}
			for _, m := range metrics {
				if m.GetCounter().GetValue() != 1 {
					t.Errorf("expected counter value 1, got %f", m.GetCounter().GetValue())
				}
			}
		}
	}
	if !totalFound {
		t.Error("requests_total metric not found")
	}
}

func TestMetricsMiddleware_SkipsHealthCheck(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.SkipHealthCheck = true

	mw, registry := MetricsMiddleware(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	var called bool
	mw(ctx, func(c *core.Ctx) error {
		called = true
		return nil
	})

	if !called {
		t.Error("health check should not be skipped")
	}

	mfs, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}

	for _, mf := range mfs {
		if mf.GetName() == "gofault_http_requests_total" {
			metrics := mf.GetMetric()
			for _, m := range metrics {
				for _, label := range m.GetLabel() {
					if label.GetName() == "path" && label.GetValue() == "/health" {
						t.Error("/health should not be recorded")
					}
				}
			}
		}
	}
}

func TestMetricsMiddleware_CustomNamespace(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.Namespace = "myapp"
	cfg.Subsystem = "api"

	mw, registry := MetricsMiddleware(cfg)
	if mw == nil {
		t.Fatal("middleware is nil")
	}

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	mw(ctx, func(c *core.Ctx) error {
		c.StatusCode = http.StatusOK
		return nil
	})

	mfs, err := registry.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}

	for _, mf := range mfs {
		name := mf.GetName()
		// Should have requests_total, request_duration_seconds, and requests_in_flight
		if name != "myapp_api_requests_total" && name != "myapp_api_request_duration_seconds" && name != "myapp_api_requests_in_flight" {
			t.Errorf("unexpected metric name: %s", name)
		}
	}
}

func TestMetricsMiddleware_InflightGauge(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.SkipHealthCheck = false

	mw, registry := MetricsMiddleware(cfg)

	req := httptest.NewRequest("GET", "/slow", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	var start time.Time
	doneCh := make(chan struct{})

	mw(ctx, func(c *core.Ctx) error {
		start = time.Now()
		go func() {
			time.Sleep(20 * time.Millisecond)
			close(doneCh)
		}()
		<-doneCh
		c.StatusCode = http.StatusOK
		return nil
	})

	// While inside handler, in-flight should be 1
	elapsed := time.Since(start)
	if elapsed < 20*time.Millisecond {
		// We're still in the handler - check gauge
		mfs, _ := registry.Gather()
		for _, mf := range mfs {
			if mf.GetName() == "gofault_http_requests_in_flight" {
				val := mf.GetMetric()[0].GetGauge().GetValue()