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
				return next(ctx)
			}
			reqPath = strings.TrimPrefix(reqPath, config.Prefix)
			// Ensure we have leading slash
			if !strings.HasPrefix(reqPath, "/") {
				reqPath = "/" + reqPath
			}
		}

		// Resolve file path
		filePath := filepath.Join(config.Dir, filepath.Clean(reqPath))

		// Security: ensure file path is within root directory. A plain
		// HasPrefix check would also accept sibling directories that merely
		// share the root's name (root /var/www would match /var/www-secret),
		// so the root is compared with a trailing separator.
		absDir, err := filepath.Abs(config.Dir)
		if err != nil {
			return next(ctx)
		}
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return next(ctx)
		}
		if !isWithinRoot(absDir, absPath) {
			return next(ctx)
		}

		// A lexical check cannot see through a symlink: <root>/link -> /etc
		// passes isWithinRoot yet resolves outside the root. os.Stat below
		// follows links, so the real target has to be validated too unless
		// following links was explicitly requested.
		if !config.FollowSymLinks {
			resolved, err := filepath.EvalSymlinks(absPath)
			if err != nil {
				// Broken or unresolvable link: nothing to serve.
				return next(ctx)
			}
			resolvedDir, err := filepath.EvalSymlinks(absDir)
			if err != nil {
				resolvedDir = absDir
			}
			if !isWithinRoot(resolvedDir, resolved) {
				return next(ctx)
			}
		}

		// Check if path is a directory
		info, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return next(ctx)
			}
			return err
		}

		// Handle directory
		if info.IsDir() {
			// Look for index file
			indexPath := filepath.Join(absPath, config.Index)
			if stat, err := os.Stat(indexPath); err == nil && !stat.IsDir() {
				absPath = indexPath
				filePath = filepath.Join(filePath, config.Index)
				info = stat
			} else if config.Browse {
				// Serve directory listing
				entries, err := os.ReadDir(absPath)
				if err != nil {
					return err
				}
				html := DirectoryListing(entries, reqPath)
				ctx.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
				ctx.Response.WriteHeader(http.StatusOK)
				ctx.Response.Write([]byte(html))
				return nil
			} else {
				// No index and browsing disabled
				return next(ctx)
			}
		}

		// Serve the file
		return serveFile(ctx, absPath, info, config)
	}
}

// isWithinRoot reports whether target is root itself or lives underneath it.
// Both paths must be absolute and already cleaned. The separator is appended
// to root so that a sibling directory sharing the root's name prefix is
// rejected rather than served.
func isWithinRoot(root, target string) bool {
	if root == target {
		return true
	}
	withSep := root
	if !strings.HasSuffix(withSep, string(os.PathSeparator)) {
		withSep += string(os.PathSeparator)
	}
	return strings.HasPrefix(target, withSep)
}

// serveFile writes the file to the response.
func serveFile(ctx *core.Ctx, absPath string, info os.FileInfo, config StaticConfig) error {
	// Set cache control if configured
	if config.CacheControl != "" {
		ctx.Response.Header().Set("Cache-Control", config.CacheControl)
	}

	// Set content type
	contentType := detectContentType(absPath, config)
	if contentType != "" {
		ctx.Response.Header().Set("Content-Type", contentType)
	}

	// Set content length
	ctx.Response.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

	// Handle range requests
	if rangeHeader := ctx.Request.Header.Get("Range"); rangeHeader != "" {
		return serveRangeRequest(ctx, absPath, info, rangeHeader)