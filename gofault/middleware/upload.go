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