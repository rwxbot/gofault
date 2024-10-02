package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultRecoveryConfig(t *testing.T) {
	config := DefaultRecoveryConfig()
	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.StackTrace {
		t.Error("Expected StackTrace to be false")
	}
}

func TestRecoveryMiddleware_PanicRecovered(t *testing.T) {
	middleware := RecoveryMiddleware(DefaultRecoveryConfig())