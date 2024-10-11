// Package middleware provides common HTTP middleware for gofault.
package middleware

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/gofault/gofault/core"
)
