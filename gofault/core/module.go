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