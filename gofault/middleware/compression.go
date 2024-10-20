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
func CompressionMiddleware(config CompressionConfig) core.MiddlewareFunc {
	if !config.Enabled {
		return nil
	}

	level := config.Level
	if level < gzip.DefaultCompression || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	minSize := config.MinSize
	if minSize < 0 {
		minSize = 0
	}

	return func(ctx *core.Ctx, next core.Handler) error {
		encoding := ctx.Request.Header.Get("Accept-Encoding")

		if !strings.Contains(encoding, "gzip") {
			return next(ctx)
		}

		// Capture the response for compression
		capture := &compressCapture{
			ResponseWriter: ctx.Response,
			statusCode:     http.StatusOK,
			body:           []byte{},
		}
		ctx.Response = capture

		err := next(ctx)

		// Only compress if body is large enough
		if len(capture.body) >= minSize {
			gz, gzipErr := gzip.NewWriterLevel(capture.ResponseWriter, level)
			if gzipErr != nil {
				// Fallback to uncompressed
				capture.ResponseWriter.WriteHeader(capture.statusCode)
				capture.ResponseWriter.Write(capture.body)
				return err
			}

			capture.ResponseWriter.Header().Set("Content-Encoding", "gzip")
			capture.ResponseWriter.Header().Set("Vary", "Accept-Encoding")
			capture.ResponseWriter.Header().Del("Content-Length")

			capture.ResponseWriter.WriteHeader(capture.statusCode)
			gz.Write(capture.body)
			gz.Close()
		} else {
			// Body too small, write uncompressed
			capture.ResponseWriter.WriteHeader(capture.statusCode)
			capture.ResponseWriter.Write(capture.body)
		}

		return err
	}
}

// compressCapture captures the response body for potential compression.
type compressCapture struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (r *compressCapture) WriteHeader(statusCode int) {
	r.statusCode = statusCode
}

func (r *compressCapture) Write(data []byte) (int, error) {
	r.body = append(r.body, data...)
	return len(data), nil
}
