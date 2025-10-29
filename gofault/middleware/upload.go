package middleware

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gofault/gofault/core"
)

// UploadConfig holds configuration for the upload middleware.
type UploadConfig struct {
	// Enabled enables the upload middleware.
	Enabled bool
	// MaxSize is the maximum file size in bytes (0 = unlimited).
	MaxSize int64
	// AllowedTypes is a list of allowed MIME types. If empty, all types are allowed.
	AllowedTypes []string
	// AllowedExtensions is a list of allowed file extensions (case-insensitive).
	AllowedExtensions []string
	// Storage is the storage backend for uploaded files.
	Storage StorageBackend
	// FieldName is the form field name for file uploads (default: "file").
	FieldName string
	// SkipFunc returns true to skip uploading for a given request.
	SkipFunc func(*core.Ctx) bool
}

// DefaultUploadConfig returns a default upload configuration.
func DefaultUploadConfig() UploadConfig {
	return UploadConfig{
		Enabled:           true,
		MaxSize:           10 * 1024 * 1024, // 10MB
		AllowedTypes:      nil,
		AllowedExtensions: nil,
		Storage:           nil, // must be set by caller
		FieldName:         "file",
		SkipFunc:          nil,
	}
}

// FileInfo holds information about an uploaded file.
type FileInfo struct {
	// FieldName is the form field name.
	FieldName string
	// FileName is the original file name.
	FileName string
	// Size is the file size in bytes.
	Size int64