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
