package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

func dummyHandler(ctx *core.Ctx) error {
	return nil
}

func TestRouter_HandleAndServe(t *testing.T) {
	r := New()
	r.Handle("GET", "/hello/:name", dummyHandler)

	req := httptest.NewRequest("GET", "/hello/world", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRouter_NotFound(t *testing.T) {
	r := New()
	r.Handle("GET", "/hello/:name", dummyHandler)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestRouter_MiddlewareChain(t *testing.T) {
	r := New()
	called := false
	r.Middleware(func(ctx *core.Ctx, next core.Handler) error {
		called = true
		return next(ctx)
	})
	r.Handle("GET", "/test", dummyHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !called {
		t.Fatal("middleware was not called")