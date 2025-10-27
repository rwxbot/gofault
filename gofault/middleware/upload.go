package middleware

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofault/gofault/core"
)

// UploadConfig holds configuration for the upload middleware.
type UploadConfig struct {
	// Enabled enables the upload middleware.
	Enabled bool
	// MaxSize is the maximum file size in bytes (0 = unlimited).
	MaxSize int64