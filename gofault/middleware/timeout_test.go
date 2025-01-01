package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofault/gofault/core"
)

func TestDefaultTimeoutConfig(t *testing.T) {
	config := DefaultTimeoutConfig()
	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.Duration != 30*time.Second {
		t.Errorf("Expected Duration 30s, got %v", config.Duration)
	}
}

func TestTimeoutMiddleware_Disabled(t *testing.T) {
	config := DefaultTimeoutConfig()
	config.Enabled = false

	handler := TimeoutMiddleware(config)
	if handler != nil {
		t.Error("Expected nil handler when disabled")
	}
}
