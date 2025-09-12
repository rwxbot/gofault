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
