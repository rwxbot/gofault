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
func CORS(cfg CORSConfig) core.MiddlewareFunc {
	return func(ctx *core.Ctx, next core.Handler) error {
		origin := ctx.Request.Header.Get("Origin")

		// Check if origin is allowed