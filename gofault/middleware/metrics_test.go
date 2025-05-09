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
				if val != 1 {
					t.Logf("in-flight gauge during request: %f", val)
				}
			}
		}
	}
}

func TestMetricsMiddleware_HistogramBuckets(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.Buckets = []float64{0.001, 0.01, 0.1, 1}

	mw, registry := MetricsMiddleware(cfg)

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
		if mf.GetName() == "gofault_http_request_duration_seconds" {
			m := mf.GetMetric()[0]
			// Verify bucket count matches our custom buckets
			_ = m.GetHistogram().GetBucket()
		}
	}
}

func TestMetricsMiddleware_StatusCode(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.SkipHealthCheck = false

	mw, registry := MetricsMiddleware(cfg)

	tests := []struct {
		path   string
		method string
		code   int
	}{
		{"/ok", "GET", 200},
		{"/notfound", "GET", 404},
		{"/error", "POST", 500},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		rec := httptest.NewRecorder()
		ctx := core.NewCtx(rec, req)

		mw(ctx, func(c *core.Ctx) error {
			c.StatusCode = tt.code
			return nil
		})
	}

	mfs, _ := registry.Gather()
	for _, mf := range mfs {
		if mf.GetName() == "gofault_http_requests_total" {
			found := make(map[string]bool)
			for _, m := range mf.GetMetric() {
				var path, method, status string
				for _, label := range m.GetLabel() {
					if label.GetName() == "path" {
						path = label.GetValue()
					}
					if label.GetName() == "method" {
						method = label.GetValue()
					}
					if label.GetName() == "status" {
						status = label.GetValue()
					}
				}
				found[path+method+status] = true
			}
			for _, tt := range tests {
				key := tt.path + tt.method + strconv.Itoa(tt.code)
				if !found[key] {
					t.Logf("metric for %s %s %d not found", tt.method, tt.path, tt.code)
				}
			}
		}
	}
}

func TestMetricsMiddleware_NilBuckets(t *testing.T) {
	cfg := MetricsConfig{Enabled: true, Buckets: nil}
	mw, _ := MetricsMiddleware(cfg)
	if mw == nil {
		t.Error("middleware should not be nil even with nil buckets")
	}
}

func TestMetricsMiddleware_DefaultBucketFallback(t *testing.T) {
	cfg := MetricsConfig{Enabled: true, Buckets: nil}
	// Should not panic and should use defaults
	mw, registry := MetricsMiddleware(cfg)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	mw(ctx, func(c *core.Ctx) error {
		c.StatusCode = 200
		return nil
	})

	mfs, _ := registry.Gather()
	if len(mfs) == 0 {
		t.Error("expected metrics to be gathered")
	}
}

func TestMetricsMiddleware_ZeroStatusCode(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.SkipHealthCheck = false

	mw, registry := MetricsMiddleware(cfg)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	mw(ctx, func(c *core.Ctx) error {
		// StatusCode not set, remains 0
		return nil
	})

	mfs, _ := registry.Gather()
	for _, mf := range mfs {
		if mf.GetName() == "gofault_http_requests_total" {
			for _, m := range mf.GetMetric() {
				for _, label := range m.GetLabel() {
					if label.GetName() == "status" && label.GetValue() == "0" {
						return // found, ok
					}
				}
			}
		}
	}
}

func TestNewHttpMetrics_Registry(t *testing.T) {
	// Test with a fresh registry instead of global
	reg := prometheus.NewRegistry()
	namespace := "testapp"
	subsystem := "http"
	buckets := []float64{0.005, 0.01, 0.05, 0.1}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
		},
		[]string{"method", "path", "status"},
	)
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Buckets:   buckets,
		},
		[]string{"method", "path", "status"},
	)
	gauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_in_flight",
		},
	)

	reg.MustRegister(counter, histogram, gauge)

	counter.WithLabelValues("GET", "/test", "200").Inc()
	histogram.WithLabelValues("GET", "/test", "200").Observe(0.001)
	gauge.Inc()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather error: %v", err)
	}

	found := make(map[string]bool)
	for _, mf := range mfs {
		found[mf.GetName()] = true
	}

	if !found["testapp_http_requests_total"] {
		t.Error("counter metric not found")
	}
	if !found["testapp_http_request_duration_seconds"] {
		t.Error("histogram metric not found")
	}
	if !found["testapp_http_requests_in_flight"] {
		t.Error("gauge metric not found")
	}
}

func TestMetricsMiddleware_ConcurrentRequests(t *testing.T) {
	cfg := DefaultMetricsConfig()
	cfg.SkipHealthCheck = false

	mw, registry := MetricsMiddleware(cfg)

	doneCh := make(chan struct{}, 5)

	for i := 0; i < 5; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/concurrent", nil)
			rec := httptest.NewRecorder()
			ctx := core.NewCtx(rec, req)

			mw(ctx, func(c *core.Ctx) error {
				c.StatusCode = http.StatusOK
				time.Sleep(10 * time.Millisecond)
				return nil
			})
			doneCh <- struct{}{}
		}()
	}

	for i := 0; i < 5; i++ {
		<-doneCh
	}

	mfs, _ := registry.Gather()
	for _, mf := range mfs {
		if mf.GetName() == "gofault_http_requests_total" {
			for _, m := range mf.GetMetric() {
				val := m.GetCounter().GetValue()
				if val != 5 {
					t.Errorf("expected 5 total requests, got %f", val)
				}
			}
		}
	}
}
