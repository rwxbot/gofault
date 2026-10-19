// Package middleware provides common HTTP middleware for gofault.
package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gofault/gofault/core"
)

// CORSConfig holds CORS configuration options.
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns the default CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORS creates a CORS middleware with the given configuration.
//
// The Access-Control-Allow-Origin header is chosen per the Fetch spec:
// a literal "*" is only legal without credentials. With AllowCredentials
// the header must name the exact request origin, otherwise every browser
// rejects the response — the combination "*" + "credentials: true" is
// invalid, not a permissive one. Whenever the answer depends on the
// request's Origin, the response carries Vary: Origin so a shared cache
// cannot serve one origin's answer to another.
func CORS(cfg CORSConfig) core.MiddlewareFunc {
	return func(ctx *core.Ctx, next core.Handler) error {
		origin := ctx.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowedOrigin := ""
		exactMatch := false
		for _, o := range cfg.AllowOrigins {
			if o == "*" {
				allowedOrigin = "*"
				break
			}
			if o == origin {
				allowedOrigin = origin
				exactMatch = true
				break
			}
			// Support wildcard subdomain matching: *.example.com matches
			// sub.example.com but NOT notexample.com or evil-example.com.
			// The suffix must be preceded by a dot.
			if strings.HasPrefix(o, "*.") {
				domain := strings.TrimPrefix(o, "*.")
				if strings.HasSuffix(origin, "."+domain) {
					allowedOrigin = origin
					exactMatch = true
					break
				}
			}
		}

		if allowedOrigin != "" {
			if allowedOrigin == "*" && cfg.AllowCredentials {
				// "*" with credentials is rejected by every browser, so the
				// configured wildcard degrades to echoing the request's
				// origin. The response is now origin-specific: Vary: Origin
				// below is what keeps a shared cache honest about that.
				// A request without an Origin header is not a CORS request;
				// echoing an empty origin would only produce a header no
				// browser accepts, so the whole block is skipped.
				allowedOrigin = origin
				exactMatch = true
			}
			if allowedOrigin != "" {
				ctx.Response.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				if allowedOrigin != "*" || exactMatch {
					ctx.Response.Header().Add("Vary", "Origin")
				}
				if cfg.AllowCredentials {
					ctx.Response.Header().Set("Access-Control-Allow-Credentials", "true")
				}
				if len(cfg.ExposeHeaders) > 0 {
					ctx.Response.Header().Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposeHeaders, ", "))
				}
				if cfg.MaxAge > 0 {
					ctx.Response.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
				}
			}
		}

		// Handle preflight request
		if ctx.Request.Method == "OPTIONS" {
			if len(cfg.AllowMethods) > 0 {
				ctx.Response.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowMethods, ", "))
			}
			if len(cfg.AllowHeaders) > 0 {
				ctx.Response.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowHeaders, ", "))
			}
			ctx.StatusCode = http.StatusNoContent
			ctx.Response.WriteHeader(http.StatusNoContent)
			return nil
		}

		return next(ctx)
	}
}
