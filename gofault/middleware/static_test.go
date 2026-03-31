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