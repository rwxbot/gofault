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
	Enabled bool
	// Level is the gzip compression level (1-9, default 6).
	Level int
	// MinSize is the minimum response size to compress (bytes).
	MinSize int
}

// DefaultCompressionConfig returns a default compression configuration.
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Enabled: true,
		Level:   gzip.DefaultCompression,
		MinSize: 1024, // 1KB minimum
	}
}

// gzipResponseWriter wraps http.ResponseWriter to compress output with gzip.
type gzipResponseWriter struct {
	http.ResponseWriter
	writer *gzip.Writer
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	return w.writer.Write(data)
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

// CompressionMiddleware compresses responses using gzip when Accept-Encoding contains "gzip".
// Returns a MiddlewareFunc that wraps the response if the client supports gzip.