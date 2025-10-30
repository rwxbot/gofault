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
	// ContentType is the MIME type.
	ContentType string
	// StoredPath is the path where the file was stored.
	StoredPath string
	// Extension is the file extension.
	Extension string
}

// StorageBackend is the interface for file storage backends.
type StorageBackend interface {
	// Store saves a file and returns its stored path.
	Store(file *multipart.FileHeader, dir string) (string, error)
	// Delete removes a stored file.
	Delete(path string) error
}

// LocalStorage implements StorageBackend using the local filesystem.
type LocalStorage struct {
	// BaseDir is the base directory for storing files.
	BaseDir string
	// Permissions for created directories and files.
	Perm os.FileMode
}

// NewLocalStorage creates a new LocalStorage backend.