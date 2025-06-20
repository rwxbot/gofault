// Package middleware provides common HTTP middleware for gofault.
package middleware

import (
	"strconv"
	"strings"

	"github.com/gofault/gofault/core"
)

// CORSConfig holds CORS configuration options.