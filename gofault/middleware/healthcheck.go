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
		config.Path = "/health"
	}
	if config.Checks == nil {
		config.Checks = make(map[string]HealthChecker)
	}

	return func(ctx *core.Ctx) error {
		response := HealthCheckResponse{
			Status: HealthStatusUp,
		}

		if config.IncludeTimestamp {
			response.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}

		if config.IncludeChecks && len(config.Checks) > 0 {
			response.Checks = make(map[string]HealthCheck)
			for name, check := range config.Checks {
				result := check()
				response.Checks[name] = result
				if result.Status == HealthStatusDown {
					response.Status = HealthStatusDown
				}
			}
		}

		statusCode := http.StatusOK
		if response.Status == HealthStatusDown {
			statusCode = http.StatusServiceUnavailable
		}

		ctx.Response.Header().Set("Content-Type", "application/json")
		ctx.Response.WriteHeader(statusCode)

		encoder := json.NewEncoder(ctx.Response)
		return encoder.Encode(response)
	}
}

// RegisterHealthCheck registers a new health check.
func (c *HealthCheckConfig) RegisterHealthCheck(name string, checker HealthChecker) {
	if c.Checks == nil {
		c.Checks = make(map[string]HealthChecker)
	}
	c.Checks[name] = checker
}

// NewHealthCheckResponse creates a basic health check response.
func NewHealthCheckResponse(status HealthStatus, message string) HealthCheck {
	return HealthCheck{
		Status:  status,
		Message: message,
	}
}

// LatencyHealthCheck creates a health check that measures latency.
func LatencyHealthCheck(name string, fn func() error) HealthChecker {
	return func() HealthCheck {
		start := time.Now()
		err := fn()
		latency := time.Since(start)

		if err != nil {
			return HealthCheck{
				Status:  HealthStatusDown,
				Message: err.Error(),
				Latency: latency.String(),
			}
		}

		return HealthCheck{
			Status:  HealthStatusUp,
			Latency: latency.String(),
		}
	}
}
