package middleware

import (
	"strconv"
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