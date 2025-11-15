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