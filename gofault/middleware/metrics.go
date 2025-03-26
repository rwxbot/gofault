package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsConfig holds configuration for the metrics middleware.
type MetricsConfig struct {
	// Enabled enables the metrics middleware.
	Enabled bool
	// Namespace prefixes all metric names.
	Namespace string
	// Subsystem further qualifies the metric.
	Subsystem string
	// SkipHealthCheck excludes the /health path from metrics.
	SkipHealthCheck bool
	// Buckets for the request duration histogram.
	Buckets []float64
}

// DefaultMetricsConfig returns a default metrics configuration.
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Enabled:         true,
		Namespace:       "gofault",
		Subsystem:       "http",
		SkipHealthCheck: true,
		Buckets:         []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}