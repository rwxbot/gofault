// Package core defines the fundamental interfaces for the gofault framework.
// All core interfaces are defined here to avoid circular dependencies.
package core

import (
	"context"
	"net/http"
)

// Context wraps the standard HTTP request context with framework-specific data.
type Context = context.Context

// Ctx encapsulates an HTTP request/response pair plus extracted route parameters.
type Ctx struct {
	Request    *http.Request
	Response   http.ResponseWriter
	Params     map[string]string
	StatusCode int
	// Locals holds arbitrary context values, used to pass data between middleware.
	// For example, WebSocket middleware stores the upgraded connection here.
	Locals map[string]any
}

func NewCtx(w http.ResponseWriter, r *http.Request) *Ctx {
	return &Ctx{Request: r, Response: w, Params: make(map[string]string), StatusCode: http.StatusOK, Locals: make(map[string]any)}
}

// SetLocal stores a value in Locals, creating the map when the Ctx was built
// without NewCtx. Middleware should use this instead of assigning into Locals
// directly, so a minimally-constructed Ctx cannot panic on a nil map.
func (c *Ctx) SetLocal(key string, value any) {
	if c.Locals == nil {
		c.Locals = make(map[string]any)
	}
	c.Locals[key] = value
}

// GetVersion returns the API version extracted by versioning middleware, or 0 if not set.
func (c *Ctx) GetVersion() int {
	if v, ok := c.Locals["version"].(int); ok {
		return v