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
	}
}

func TestStaticMiddleware_NotFound_PassesToNext(t *testing.T) {
	tmpDir := t.TempDir()
	config := StaticConfig{Dir: tmpDir}
	mw := StaticMiddleware(config)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	nextCalled := false
	_ = mw(ctx, func(ctx *core.Ctx) error {
		nextCalled = true
		ctx.Response.WriteHeader(http.StatusNotFound)
		return nil
	})

	if !nextCalled {
		t.Error("next handler should be called for non-existent file")
	}
}

func TestStaticMiddleware_Prefix_Stripped(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("prefixed"), 0644)

	config := StaticConfig{
		Dir:    tmpDir,
		Prefix: "/static",
	}

	mw := StaticMiddleware(config)

	req := httptest.NewRequest("GET", "/static/file.txt", nil)
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

	if w.Body.String() != "prefixed" {
		t.Errorf("expected 'prefixed', got '%s'", w.Body.String())
	}
}

func TestStaticMiddleware_DirectoryListing(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "mydir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "mydir", "file.txt"), []byte("content"), 0644)

	config := StaticConfig{
		Dir:    tmpDir,
		Browse: true,
	}

	mw := StaticMiddleware(config)

	req := httptest.NewRequest("GET", "/mydir", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := mw(ctx, func(ctx *core.Ctx) error {
		t.Error("next should not be called for directory with Browse enabled")
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Index of /mydir") {
		t.Errorf("expected directory listing, got '%s'", body)
	}

	if !strings.Contains(body, "file.txt") {
		t.Errorf("expected file.txt in listing, got '%s'", body)
	}

	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/html; charset=utf-8', got '%s'", ct)
	}
}

func TestStaticMiddleware_NoCacheControl(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("nocache"), 0644)

	config := StaticConfig{
		Dir:          tmpDir,
		CacheControl: "", // disabled
	}

	mw := StaticMiddleware(config)

	req := httptest.NewRequest("GET", "/file.txt", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := mw(ctx, func(ctx *core.Ctx) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cc := w.Header().Get("Cache-Control"); cc != "" {
		t.Errorf("expected no Cache-Control, got '%s'", cc)
	}
}

func TestStaticMiddleware_RangeRequest(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "range.txt"), []byte("0123456789"), 0644)

	config := StaticConfig{Dir: tmpDir}
	mw := StaticMiddleware(config)

	req := httptest.NewRequest("GET", "/range.txt", nil)
	req.Header.Set("Range", "bytes=0-4")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := mw(ctx, func(ctx *core.Ctx) error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Code != http.StatusPartialContent {
		t.Errorf("expected status 206, got %d", w.Code)
	}

	if cr := w.Header().Get("Content-Range"); cr != "0-4/10" {
		t.Errorf("expected Content-Range '0-4/10', got '%s'", cr)
	}

	if w.Body.String() != "01234" {
		t.Errorf("expected '01234', got '%s'", w.Body.String())
	}
}

func TestStaticMiddleware_RangeRequest_Suffix(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "suffix.txt"), []byte("0123456789"), 0644)

	config := StaticConfig{Dir: tmpDir}
	mw := StaticMiddleware(config)

	req := httptest.NewRequest("GET", "/suffix.txt", nil)
	req.Header.Set("Range", "bytes=-5")