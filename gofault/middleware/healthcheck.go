package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
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
//
// Checks is guarded by a mutex: the handler reads it on every scrape while
// RegisterHealthCheck may be called at runtime, and a concurrent map read and
// write panics.
//
// The mutex is a value, not a pointer, so it needs no initialisation. A
// pointer would have to be installed lazily, and that install is itself a race
// that no amount of locking on the read side can fix. It also means copying a
// config is now a bug that go vet reports, which is correct: a copy would get a
// mutex of its own and provide no protection over the shared map.
//
// Because a config must not be copied, the constructors and handlers take
// *HealthCheckConfig.
type HealthCheckConfig struct {
	// Path is the URL path for the health check endpoint.
	Path string
	// Checks is a map of health checkers. Assign to it directly only before
	// the handler starts serving; use RegisterHealthCheck afterwards.
	Checks map[string]HealthChecker
	// IncludeTimestamp controls whether to include timestamp in response.
	IncludeTimestamp bool
	// IncludeChecks controls whether to include the individual check details.
	IncludeChecks bool

	mu sync.RWMutex
}

// snapshot returns the checkers, read under the lock.
func (c *HealthCheckConfig) snapshot() map[string]HealthChecker {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.Checks) == 0 {
		return nil
	}
	out := make(map[string]HealthChecker, len(c.Checks))
	for name, check := range c.Checks {
		out[name] = check
	}
	return out
}

// RegisterHealthCheck registers a new health check.
func (c *HealthCheckConfig) RegisterHealthCheck(name string, checker HealthChecker) {
	if checker == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Checks == nil {
		c.Checks = make(map[string]HealthChecker)
	}
	c.Checks[name] = checker
}

// DefaultHealthCheckConfig returns a default health check configuration.
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		Path:             "/health",
		Checks:           make(map[string]HealthChecker),
		IncludeTimestamp: true,
		IncludeChecks:    true,
	}
}

// HealthCheckHandler creates the health check endpoint handler.
//
// It returns a handler rather than a middleware: the health endpoint is a
// terminal route, so RegisterHealthCheck wires the path for you.
func HealthCheckHandler(config *HealthCheckConfig) core.Handler {
	if config.Path == "" {
		config.Path = "/health"
	}

	return func(ctx *core.Ctx) error {
		response := HealthCheckResponse{
			Status: HealthStatusUp,
		}

		if config.IncludeTimestamp {
			response.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}

		checks := config.snapshot()
		if config.IncludeChecks && len(checks) > 0 {
			response.Checks = make(map[string]HealthCheck, len(checks))
			for name, check := range checks {
				// A panicking checker must not take down the health endpoint;
				// report it as down and keep the other results.
				result := runChecker(check)
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