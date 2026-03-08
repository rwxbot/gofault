package middleware

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofault/gofault/core"
)

// StaticConfig holds configuration for the static file middleware.
type StaticConfig struct {
	// Dir is the root directory to serve files from.
	Dir string
	// Prefix is the URL path prefix to strip before looking up files.
	Prefix string
	// Index is the default index file name.
	Index string
	// CacheControl sets the Cache-Control header value. Empty string disables.
	CacheControl string
	// Browse enables directory listing.
	Browse bool
	// FollowSymLinks follows symbolic links.
	FollowSymLinks bool
	// ExtraExtensions maps file extensions to custom MIME types.
	ExtraExtensions map[string]string
}

// DefaultStaticConfig returns a default static file configuration.
func DefaultStaticConfig() StaticConfig {
	return StaticConfig{
		Index:          "index.html",
		CacheControl:   "public, max-age=3600",
		Browse:         false,
		FollowSymLinks: false,
	}
}

// StaticMiddleware creates a middleware that serves static files.
func StaticMiddleware(config StaticConfig) core.MiddlewareFunc {
	if config.Index == "" {
		config.Index = "index.html"
	}

	return func(ctx *core.Ctx, next core.Handler) error {
		reqPath := ctx.Request.URL.Path

		// Strip prefix if configured
		if config.Prefix != "" {
			if !strings.HasPrefix(reqPath, config.Prefix) {