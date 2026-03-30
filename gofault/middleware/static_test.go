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