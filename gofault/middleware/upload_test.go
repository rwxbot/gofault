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