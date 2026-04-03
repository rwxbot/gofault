package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestStaticMiddleware_ServeFile(t *testing.T) {
	// Create temp directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("Hello, World!"), 0644)

	config := StaticConfig{
		Dir:          tmpDir,
		CacheControl: "public, max-age=3600",
	}

	mw := StaticMiddleware(config)

	// Create a request
	req := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := mw(ctx, func(ctx *core.Ctx) error {
		t.Error("next handler should not be called for existing file")
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "Hello, World!" {
		t.Errorf("expected body 'Hello, World!', got '%s'", w.Body.String())
	}

	if ct := w.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/plain; charset=utf-8', got '%s'", ct)
	}

	if cc := w.Header().Get("Cache-Control"); cc != "public, max-age=3600" {
		t.Errorf("expected Cache-Control 'public, max-age=3600', got '%s'", cc)
	}
}

func TestStaticMiddleware_DirectoryIndex(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "index.html"), []byte("<h1>Index</h1>"), 0644)

	config := StaticConfig{
		Dir:   tmpDir,
		Index: "index.html",
	}

	mw := StaticMiddleware(config)

	req := httptest.NewRequest("GET", "/subdir", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := mw(ctx, func(ctx *core.Ctx) error {
		t.Error("next should not be called")
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "<h1>Index</h1>") {
		t.Errorf("expected index content, got '%s'", w.Body.String())