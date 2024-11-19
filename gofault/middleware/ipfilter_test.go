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
}

func TestIPFilterMiddleware_Disabled(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Enabled = false

	handler := IPFilterMiddleware(config)
	if handler != nil {
		t.Error("Expected nil handler when disabled")
	}
}

func TestIPFilterMiddleware_NoRestrictions(t *testing.T) {
	middleware := IPFilterMiddleware(DefaultIPFilterConfig())

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)