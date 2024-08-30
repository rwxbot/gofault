package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofault/gofault/core"
)

// HealthStatus represents the health status of a component.
type HealthStatus string

const (
	HealthStatusUp   HealthStatus = "up"
	HealthStatusDown HealthStatus = "down"
)

// HealthCheckResponse represents the health check response.
type HealthCheckResponse struct {
	Status    HealthStatus           `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Checks    map[string]HealthCheck `json:"checks,omitempty"`
}

// HealthCheck represents a single health check.
type HealthCheck struct {
	Status  HealthStatus `json:"status"`
	Message string       `json:"message,omitempty"`
	Latency string       `json:"latency,omitempty"`
}

// HealthChecker is a function that performs a health check.
type HealthChecker func() HealthCheck

// HealthCheckConfig holds configuration for the health check middleware.
type HealthCheckConfig struct {
	// Path is the URL path for the health check endpoint.
	Path string
	// Checks is a map of health checkers.
	Checks map[string]HealthChecker
	// IncludeTimestamp controls whether to include timestamp in response.
	IncludeTimestamp bool
	// IncludeChecks controls whether to include individual check details.
	IncludeChecks bool
}

// DefaultHealthCheckConfig returns a default health check configuration.
func DefaultHealthCheckConfig() HealthCheckConfig {
	return HealthCheckConfig{
		Path:             "/health",
		Checks:           make(map[string]HealthChecker),
		IncludeTimestamp: true,
		IncludeChecks:    true,
	}
}

// HealthCheckMiddleware creates a health check endpoint.
func HealthCheckMiddleware(config HealthCheckConfig) core.Handler {
	if config.Path == "" {