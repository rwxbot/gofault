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