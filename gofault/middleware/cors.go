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
					break
				}
			}
		}

		if allowedOrigin != "" {
			ctx.Response.Header().Set("Access-Control-Allow-Origin", allowedOrigin)