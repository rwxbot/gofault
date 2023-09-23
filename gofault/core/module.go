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