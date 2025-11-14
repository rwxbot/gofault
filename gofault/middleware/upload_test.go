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