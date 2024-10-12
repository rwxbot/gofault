package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gofault/gofault/core"
)

// CompressionConfig holds configuration for the compression middleware.
type CompressionConfig struct {
	// Enabled enables the compression middleware.