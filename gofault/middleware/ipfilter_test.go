package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultIPFilterConfig(t *testing.T) {
	config := DefaultIPFilterConfig()
	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.Mode != "block" {
		t.Errorf("Expected Mode 'block', got '%s'", config.Mode)
	}