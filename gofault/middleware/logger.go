// Package middleware provides common HTTP middleware for gofault.
package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"time"

	"github.com/gofault/gofault/core"
)
