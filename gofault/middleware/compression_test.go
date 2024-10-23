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
