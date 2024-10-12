// Package middleware provides common HTTP middleware for gofault.
package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gofault/gofault/core"
)

// CompressionConfig holds configuration for the compression middleware.
type CompressionConfig struct {
	// Enabled enables the compression middleware.
	Enabled bool
	// Level is the gzip compression level: gzip.HuffmanOnly (-2) through
	// gzip.BestCompression (9), which includes NoCompression (0),
	// DefaultCompression (-1) and BestSpeed (1). Out-of-range values fall
	// back to gzip.DefaultCompression (6) or clamp to the nearest bound.
	Level int
	// MinSize is the minimum response size to compress (bytes).
	MinSize int
}
