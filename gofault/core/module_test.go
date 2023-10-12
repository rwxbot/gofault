package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewCtx(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewCtx(w, r)
	if ctx.Request != r {
		t.Error("Request not set")
	}
	if ctx.Response != w {
		t.Error("Response not set")
	}
	if ctx.StatusCode != http.StatusOK {
		t.Errorf("expected StatusOK, got %d", ctx.StatusCode)
	}
	if ctx.Params == nil {
		t.Error("Params should be initialized")
	}
	if ctx.Locals == nil {
		t.Error("Locals should be initialized")
	}
}

func TestCtx_Locals(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	ctx := NewCtx(w, r)

	// Store and retrieve
	ctx.Locals["user"] = "alice"
	ctx.Locals["score"] = 42

	if v := ctx.Locals["user"]; v != "alice" {
		t.Errorf("expected alice, got %v", v)
	}
	if v := ctx.Locals["score"]; v != 42 {