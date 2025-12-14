package middleware

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultUploadConfig(t *testing.T) {
	cfg := DefaultUploadConfig()
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.MaxSize != 10*1024*1024 {
		t.Errorf("expected MaxSize 10MB, got %d", cfg.MaxSize)
	}
	if cfg.FieldName != "file" {
		t.Errorf("expected FieldName 'file', got '%s'", cfg.FieldName)
	}
}

func TestUploadMiddleware_Disabled(t *testing.T) {
	cfg := UploadConfig{Enabled: false}
	mw := UploadMiddleware(cfg)
	if mw != nil {
		t.Error("expected nil middleware when disabled")
	}
}

func TestUploadMiddleware_NoStorage(t *testing.T) {
	cfg := DefaultUploadConfig()
	cfg.Storage = nil
	mw := UploadMiddleware(cfg)
	if mw == nil {
		t.Fatal("middleware is nil")
	}

	// Create a multipart body with a file (so storage check is reached)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(part, "hello")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	err = mw(ctx, func(c *core.Ctx) error { return nil })
	if err == nil {
		t.Error("expected error when no storage configured")
	}
}

func TestLocalStorage_StoreAndDelete(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewLocalStorage(tmpDir)

	// Create a fake multipart file header
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(part, "hello world")
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	_ = req.ParseMultipartForm(0)

	file := req.MultipartForm.File["file"][0]

	path, err := storage.Store(file, "subdir")
	if err != nil {
		t.Fatalf("store error: %v", err)
	}

	expected := filepath.Join("subdir", "test.txt")
	if path != expected {
		t.Errorf("expected path %s, got %s", expected, path)
	}

	// Verify file exists
	fullPath := filepath.Join(tmpDir, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("read file error: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(data))
	}

	// Delete
	err = storage.Delete(path)
	if err != nil {
		t.Fatalf("delete error: %v", err)
	}
	if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
		t.Error("file should be deleted")
	}
}

func TestUploadMiddleware_UploadsFile(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewLocalStorage(tmpDir)

	cfg := DefaultUploadConfig()
	cfg.Storage = storage

	mw := UploadMiddleware(cfg)
	if mw == nil {
		t.Fatal("middleware is nil")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "myfile.txt")
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(part, "file content here")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	var handlerCalled bool
	err = mw(ctx, func(c *core.Ctx) error {
		handlerCalled = true
		files := GetUploadFiles(c)
		if len(files) != 1 {
			t.Errorf("expected 1 file, got %d", len(files))
		}
		if files[0].FileName != "myfile.txt" {
			t.Errorf("expected 'myfile.txt', got '%s'", files[0].FileName)
		}
		if files[0].Size != 17 {
			t.Errorf("expected size 17, got %d", files[0].Size)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("middleware error: %v", err)
	}
	if !handlerCalled {
		t.Error("next handler was not called")
	}
}

func TestUploadMiddleware_SkipNonMultipart(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewLocalStorage(tmpDir)

	cfg := DefaultUploadConfig()
	cfg.Storage = storage

	mw := UploadMiddleware(cfg)

	req := httptest.NewRequest("POST", "/upload", bytes.NewReader([]byte("hello")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	var called bool
	mw(ctx, func(c *core.Ctx) error {
		called = true
		return nil
	})

	if !called {
		t.Error("non-multipart request should skip upload handling")
	}
}

func TestUploadMiddleware_MaxSizeExceeded(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewLocalStorage(tmpDir)

	cfg := DefaultUploadConfig()
	cfg.Storage = storage
	cfg.MaxSize = 5 // 5 bytes

	mw := UploadMiddleware(cfg)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "big.txt")
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(part, "this file is bigger than 5 bytes")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	mw(ctx, func(c *core.Ctx) error { return nil })

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", rec.Code)
	}
}

func TestUploadMiddleware_AllowedExtensions(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewLocalStorage(tmpDir)

	cfg := DefaultUploadConfig()
	cfg.Storage = storage
	cfg.AllowedExtensions = []string{".txt", ".png"}

	mw := UploadMiddleware(cfg)

	// Test allowed extension
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "document.txt")
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(part, "content")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	var called bool
	mw(ctx, func(c *core.Ctx) error {
		called = true
		return nil
	})

	if !called {
		t.Error("allowed extension should not be rejected")
	}

	// Test disallowed extension
	body2 := &bytes.Buffer{}
	writer2 := multipart.NewWriter(body2)
	part2, err := writer2.CreateFormFile("file", "document.exe")
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(part2, "content")
	writer2.Close()

	req2 := httptest.NewRequest("POST", "/upload", body2)
	req2.Header.Set("Content-Type", writer2.FormDataContentType())
	rec2 := httptest.NewRecorder()
	ctx2 := core.NewCtx(rec2, req2)

	mw(ctx2, func(c *core.Ctx) error { return nil })

	if rec2.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415 for disallowed extension, got %d", rec2.Code)
	}
}

func TestUploadMiddleware_AllowedTypes(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewLocalStorage(tmpDir)

	cfg := DefaultUploadConfig()
	cfg.Storage = storage
	cfg.AllowedTypes = []string{"text/plain", "image/png"}

	mw := UploadMiddleware(cfg)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Use CreatePart to set explicit Content-Type: text/plain
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="document.txt"`))
	h.Set("Content-Type", "text/plain")
	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(part, "content")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	var called bool
	mw(ctx, func(c *core.Ctx) error {
		called = true
		return nil
	})

	if !called {
		t.Error("allowed type should not be rejected")
	}
}

func TestUploadMiddleware_SkipFunc(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewLocalStorage(tmpDir)

	cfg := DefaultUploadConfig()
	cfg.Storage = storage
	cfg.SkipFunc = func(ctx *core.Ctx) bool {
		return ctx.Request.URL.Path == "/skip"
	}

	mw := UploadMiddleware(cfg)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(part, "content")
	writer.Close()

	req := httptest.NewRequest("POST", "/skip", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	var called bool
	mw(ctx, func(c *core.Ctx) error {
		called = true
		return nil
	})

	if !called {
		t.Error("skip path should call next handler")
	}
}

func TestUploadMiddleware_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewLocalStorage(tmpDir)

	cfg := DefaultUploadConfig()
	cfg.Storage = storage

	mw := UploadMiddleware(cfg)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for i := 0; i < 3; i++ {
		part, err := writer.CreateFormFile("file", "file"+string(rune('0'+i))+".txt")
		if err != nil {
			t.Fatal(err)
		}
		io.WriteString(part, "content")
	}
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	var files []FileInfo
	mw(ctx, func(c *core.Ctx) error {
		files = GetUploadFiles(c)
		return nil
	})

	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}
}

func TestGetUploadFiles_NoFiles(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	files := GetUploadFiles(ctx)
	if files != nil {
		t.Error("expected nil when no files uploaded")
	}
}

func TestLocalStorage_MkdirAll(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewLocalStorage(tmpDir)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Use CreatePart to preserve exact filename with path separators
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="deep/nested/path.txt"`)
	h.Set("Content-Type", "text/plain")
	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(part, "nested")
	writer.Close()

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	_ = req.ParseMultipartForm(0)

	file := req.MultipartForm.File["file"][0]
	path, err := storage.Store(file, "a/b/c")
	if err != nil {
		t.Fatalf("store error: %v", err)
	}
	if path != "a/b/c/path.txt" {
		t.Errorf("unexpected path: %s", path)
	}
}
