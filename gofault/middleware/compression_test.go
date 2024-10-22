package middleware

import (
	"compress/gzip"
	"io"
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