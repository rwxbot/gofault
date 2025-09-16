package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestRequestID_GeneratesNewID(t *testing.T) {
	middleware := RequestID()

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, func(ctx *core.Ctx) error { return nil })

	respHeader := w.Header().Get(RequestIDHeader)
	if respHeader == "" {
		t.Error("X-Request-ID should be set")
	}
	if len(respHeader) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("X-Request-ID length = %d, want 32", len(respHeader))
	}