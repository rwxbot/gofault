package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestCORS_DefaultConfig(t *testing.T) {
	cfg := DefaultCORSConfig()
	if len(cfg.AllowOrigins) != 1 || cfg.AllowOrigins[0] != "*" {
		t.Errorf("default AllowOrigins should be [*], got %v", cfg.AllowOrigins)
	}
	if len(cfg.AllowMethods) != 6 {
		t.Errorf("default AllowMethods should have 6 methods, got %d", len(cfg.AllowMethods))
	}
	if cfg.MaxAge != 86400 {