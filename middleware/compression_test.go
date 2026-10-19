package middleware

import (
	"bufio"
	"compress/gzip"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultCompressionConfig(t *testing.T) {
	config := DefaultCompressionConfig()
	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.Level != gzip.DefaultCompression {
		t.Errorf("Expected Level %d, got %d", gzip.DefaultCompression, config.Level)
	}
	if config.MinSize != 1024 {
		t.Errorf("Expected MinSize 1024, got %d", config.MinSize)
	}
}

func TestCompressionMiddleware_NoAcceptEncoding(t *testing.T) {
	middleware := CompressionMiddleware(DefaultCompressionConfig())

	next := func(ctx *core.Ctx) error {
		ctx.Response.Header().Set("Content-Type", "text/plain")
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte(strings.Repeat("a", 2000)))
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Encoding") == "gzip" {
		t.Error("Expected no Content-Encoding header for missing Accept-Encoding")
	}
}

func TestCompressionMiddleware_Disabled(t *testing.T) {
	config := DefaultCompressionConfig()
	config.Enabled = false

	handler := CompressionMiddleware(config)
	if handler != nil {
		t.Error("Expected nil handler when disabled")
	}
}

func TestCompressCapture_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()

	capture := &compressCapture{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	capture.WriteHeader(http.StatusCreated)

	if capture.statusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, capture.statusCode)
	}
}

func TestCompressCapture_Write(t *testing.T) {
	w := httptest.NewRecorder()

	capture := &compressCapture{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	data := []byte("test data")
	n, err := capture.Write(data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes, got %d", len(data), n)
	}
	if capture.body.String() != "test data" {
		t.Errorf("Expected body 'test data', got '%s'", capture.body.String())
	}
}

func TestCompressionMiddleware_WithGzip(t *testing.T) {
	middleware := CompressionMiddleware(DefaultCompressionConfig())

	next := func(ctx *core.Ctx) error {
		ctx.Response.Header().Set("Content-Type", "text/plain")
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte(strings.Repeat("x", 2000)))
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding 'gzip', got '%s'", w.Header().Get("Content-Encoding"))
	}

	// Verify the body is actually gzip compressed
	reader, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzip body: %v", err)
	}

	if len(body) != 2000 {
		t.Errorf("Expected decompressed body length 2000, got %d", len(body))
	}
}

// Flush is provided so streaming handlers keep working through the buffer. It
// had no coverage, so a streaming handler could have silently broken.
func TestCompressCapture_FlushDelegates(t *testing.T) {
	rec := &flushRecorder{header: make(http.Header)}
	capture := newCompressCapture(rec)

	_, _ = capture.Write([]byte("x"))
	capture.Flush()

	if !rec.flushed {
		t.Error("Flush did not reach the underlying writer")
	}
}

// A handler holding the buffered writer as an http.Flusher must be able to
// flush without an error.
func TestCompressCapture_SatisfiesFlusher(t *testing.T) {
	var _ http.Flusher = newCompressCapture(httptest.NewRecorder())
}

type flushRecorder struct {
	header  http.Header
	flushed bool
}

func (f *flushRecorder) Header() http.Header         { return f.header }
func (f *flushRecorder) Write(b []byte) (int, error) { return len(b), nil }
func (f *flushRecorder) WriteHeader(int)             {}
func (f *flushRecorder) Flush()                      { f.flushed = true }

// Level: 1 (BestSpeed) and 0 (NoCompression) are legitimate gzip levels; the
// old clamp silently rewrote them into 6.
func TestNormalizeCompressionLevel(t *testing.T) {
	cases := []struct{ in, want int }{
		{gzip.BestSpeed, gzip.BestSpeed},
		{0, 0},
		{1, 1},
		{gzip.DefaultCompression, gzip.DefaultCompression},
		{gzip.BestCompression, gzip.BestCompression},
		{-3, gzip.DefaultCompression}, // below range: fall back to default
		{10, gzip.BestCompression},    // above range: clamp to best
	}
	for _, c := range cases {
		if got := normalizeCompressionLevel(c.in); got != c.want {
			t.Errorf("normalizeCompressionLevel(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

// net/http commits the response with an implicit 200 on Flush. Flushing the
// capture must therefore replay the buffered body onto the real writer first;
// delegating Flush alone would send a 200 status line while the handler's
// status and body were still sitting in the buffer, never to be seen.
func TestCompressCapture_FlushReplaysBody(t *testing.T) {
	rec := httptest.NewRecorder()
	capture := newCompressCapture(rec)

	capture.WriteHeader(http.StatusTeapot)
	_, _ = capture.Write([]byte("chunk1"))
	capture.Flush()

	if rec.Code != http.StatusTeapot {
		t.Errorf("status after Flush = %d, want %d", rec.Code, http.StatusTeapot)
	}
	if rec.Body.String() != "chunk1" {
		t.Errorf("body after Flush = %q, want %q", rec.Body.String(), "chunk1")
	}

	// A second Flush only replays what was written since the first.
	_, _ = capture.Write([]byte("chunk2"))
	capture.Flush()
	if rec.Body.String() != "chunk1chunk2" {
		t.Errorf("body after second Flush = %q, want %q", rec.Body.String(), "chunk1chunk2")
	}
}

// Once a streaming handler has flushed, the middleware must not gzip and
// replay the whole buffer — the flushed prefix would reach the client twice.
func TestCompressionMiddleware_FlushedResponseNotDoubleWritten(t *testing.T) {
	config := DefaultCompressionConfig()
	config.MinSize = 1
	mw := CompressionMiddleware(config)

	next := func(ctx *core.Ctx) error {
		f := ctx.Response.(http.Flusher)
		_, _ = ctx.Response.Write([]byte("first "))
		f.Flush()
		_, _ = ctx.Response.Write([]byte("second"))
		f.Flush()
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	if err := mw(ctx, next); err != nil {
		t.Fatalf("middleware: %v", err)
	}
	if got := w.Body.String(); got != "first second" {
		t.Errorf("body = %q, want %q (flushed prefix must not repeat)", got, "first second")
	}
	if ce := w.Header().Get("Content-Encoding"); ce == "gzip" {
		t.Error("a flushed response must not be gzip-compressed after the fact")
	}
}

// A WebSocket handshake under compression used to fail: the capture does not
// implement http.Hijacker (embedding does not extend the method set), and
// the upgrader type-asserts the writer to it. The upgrade request must pass
// through with the real writer untouched. The base writer here implements
// Hijacker the way net/http's does, so the assertion is meaningful.
func TestCompressionMiddleware_UpgradeRequestPassthrough(t *testing.T) {
	mw := CompressionMiddleware(DefaultCompressionConfig())

	next := func(ctx *core.Ctx) error {
		if _, ok := ctx.Response.(http.Hijacker); !ok {
			t.Error("upgrade handler cannot reach http.Hijacker under the compression middleware")
			return nil
		}
		ctx.Response.WriteHeader(http.StatusSwitchingProtocols)
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Connection", "Upgrade")
	w := &hijackRecorder{ResponseRecorder: httptest.NewRecorder()}
	ctx := core.NewCtx(w, req)

	if err := mw(ctx, next); err != nil {
		t.Fatalf("middleware: %v", err)
	}
	if ce := w.Header().Get("Content-Encoding"); ce != "" {
		t.Errorf("Content-Encoding = %q on an upgrade response, want none", ce)
	}
}

func TestIsUpgradeRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	if isUpgradeRequest(req) {
		t.Error("plain GET must not look like an upgrade")
	}
	req.Header.Set("Upgrade", "websocket")
	if isUpgradeRequest(req) {
		t.Error("Upgrade header without Connection: upgrade must not match")
	}
	req.Header.Set("Connection", "keep-alive, Upgrade")
	if !isUpgradeRequest(req) {
		t.Error("multi-token Connection with Upgrade must match")
	}
	req.Header.Set("Connection", "upgrade")
	if !isUpgradeRequest(req) {
		t.Error("lowercase upgrade token must match")
	}
}

// hijackRecorder is an httptest recorder that also implements http.Hijacker,
// standing in for net/http's ResponseWriter.
type hijackRecorder struct {
	*httptest.ResponseRecorder
}

func (h *hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("not really hijacked")
}
