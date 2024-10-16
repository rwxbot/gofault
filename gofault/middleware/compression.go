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

// DefaultCompressionConfig returns a default compression configuration.
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		Enabled: true,
		Level:   gzip.DefaultCompression,
		MinSize: 1024, // 1KB minimum
	}
}

// compressCapture buffers a handler's output so it can be compressed before
// anything reaches the client.
//
// The body is accumulated in a bytes.Buffer rather than a []byte. Repeatedly
// appending to a slice reallocates and copies everything written so far, which
// made a large response cost O(n^2); bytes.Buffer grows geometrically.
//
// The status code is captured the same way: net/http ignores every WriteHeader
// after the first, so a second call would be silently dropped and the client
// would see a different status than the handler chose.
type compressCapture struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	flushed     bool // the status line has reached the real writer
	sent        int  // bytes of body already replayed onto the real writer
	body        bytes.Buffer
}

func newCompressCapture(w http.ResponseWriter) *compressCapture {
	return &compressCapture{ResponseWriter: w, statusCode: http.StatusOK}
}

func (r *compressCapture) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.statusCode = statusCode
	r.wroteHeader = true
}

func (r *compressCapture) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	// bytes.Buffer.Write never returns an error.
	_, _ = r.body.Write(data)
	return len(data), nil
}

// Flush lets handlers that stream still flush, even though the body is
// buffered. Without it, http.Flusher is unavailable and streaming handlers
// either error or silently buffer everything.
//
// The buffered body must reach the real writer first: in net/http, Flush
// commits the response with an implicit 200, so flushing an empty buffer
// would send a 200 status line while the status the handler chose — and the
// body itself — are still sitting in the buffer and get dropped when the
// middleware replays it onto an already-committed writer.
func (r *compressCapture) Flush() {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	r.flushTo(r.ResponseWriter)
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// isUpgradeRequest reports whether the request asks to switch protocols —
// the WebSocket handshake shape: an Upgrade header plus a Connection header
// naming upgrade. net/http's own 101 handling assumes the writer reaches
// the client unbuffered.
func isUpgradeRequest(r *http.Request) bool {
	if r.Header.Get("Upgrade") == "" {
		return false
	}
	for _, token := range strings.Split(r.Header.Get("Connection"), ",") {
		if strings.EqualFold(strings.TrimSpace(token), "upgrade") {
			return true
		}
	}
	return false
}

// normalizeCompressionLevel clamps a configured gzip level into the range
// gzip.NewWriterLevel accepts: HuffmanOnly(-2) through BestCompression(9),
// which includes NoCompression(0), DefaultCompression(-1) and BestSpeed(1).
// The old check treated anything below DefaultCompression as invalid and
// silently rewrote Level: 1 (BestSpeed) into 6, contradicting the docs.
func normalizeCompressionLevel(level int) int {
	if level < gzip.HuffmanOnly {
		return gzip.DefaultCompression
	}
	if level > gzip.BestCompression {
		return gzip.BestCompression
	}
	return level
}

// CompressionMiddleware compresses responses using gzip when Accept-Encoding
// contains "gzip".
func CompressionMiddleware(config CompressionConfig) core.MiddlewareFunc {
	if !config.Enabled {
		return nil
	}

	level := normalizeCompressionLevel(config.Level)

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