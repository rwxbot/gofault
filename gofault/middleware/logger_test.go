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