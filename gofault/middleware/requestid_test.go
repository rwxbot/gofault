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
}

func TestRequestID_UsesProvidedID(t *testing.T) {
	middleware := RequestID()

	providedID := "custom-request-id-12345"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, providedID)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, func(ctx *core.Ctx) error { return nil })

	respHeader := w.Header().Get(RequestIDHeader)
	if respHeader != providedID {
		t.Errorf("X-Request-ID = %s, want %s", respHeader, providedID)
	}
}

func TestRequestID_ForwardsToNext(t *testing.T) {
	middleware := RequestID()

	nextCalled := false
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := middleware(ctx, func(ctx *core.Ctx) error {
		nextCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Error("next handler should be called")
	}
}

func TestRequestIDHeader_Constant(t *testing.T) {
	if RequestIDHeader != "X-Request-ID" {
		t.Errorf("RequestIDHeader = %s, want X-Request-ID", RequestIDHeader)
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := generateID()
		if ids[id] {
			t.Fatal("generated duplicate ID")
		}
		ids[id] = true
		if len(id) != 32 {
			t.Errorf("ID length = %d, want 32", len(id))
		}
	}
}
