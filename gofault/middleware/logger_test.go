package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultRequestLoggerConfig(t *testing.T) {
	config := DefaultRequestLoggerConfig()
	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.LogBody {
		t.Error("Expected LogBody to be false")
	}
	if config.LogHeaders {
		t.Error("Expected LogHeaders to be false")
	}
}

func TestRequestLoggerMiddleware_Disabled(t *testing.T) {
	config := DefaultRequestLoggerConfig()
	config.Enabled = false

	handler := RequestLoggerMiddleware(config)
	if handler != nil {
		t.Error("Expected nil handler when disabled")
	}
}
