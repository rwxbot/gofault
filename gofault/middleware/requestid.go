package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gofault/gofault/core"
)

const (
	// RequestIDHeader is the header name for request ID.
	RequestIDHeader = "X-Request-ID"
)

// RequestID generates a unique request ID and adds it to the context.
func RequestID() core.MiddlewareFunc {
	return func(ctx *core.Ctx, next core.Handler) error {
		requestID := ctx.Request.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = generateID()
		}

		ctx.Response.Header().Set(RequestIDHeader, requestID)

		return next(ctx)
	}
}

// generateID creates a random 16-byte hex string.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
