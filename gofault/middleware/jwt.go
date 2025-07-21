package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

// jwtAlgHS256 is the only signing algorithm this package supports.
const jwtAlgHS256 = "HS256"

// JWTConfig holds JWT authentication configuration.
type JWTConfig struct {
	Secret []byte
	// Algorithm must be HS256. Any other value is rejected at construction.
	Algorithm string
	// TokenName is the query parameter consulted for the token when
	// AllowQueryToken is enabled.
	TokenName string
	// AllowQueryToken permits reading the token from the URL query string.
	// Off by default: query strings land in access logs, Referer headers and
	// browser history, which leaks the credential.
	AllowQueryToken bool
	// RequireExpiry rejects tokens without an exp claim. On by default so a
	// token stays valid only as long as the issuer intended.
	RequireExpiry bool
}

// DefaultJWTConfig returns default JWT configuration.
func DefaultJWTConfig(secret []byte) JWTConfig {
	return JWTConfig{
		Secret:          secret,
		Algorithm:       jwtAlgHS256,
		TokenName:       "token",
		AllowQueryToken: false,
		RequireExpiry:   true,
	}
}

// Claims represents JWT claims.
type Claims struct {
	Subject   string   `json:"sub"`
	UserID    string   `json:"uid"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	Issuer    string   `json:"iss"`
}

// IsValid reports whether the claims are within their validity window.
// A zero ExpiresAt means "no expiry recorded"; callers that require a bounded
// lifetime must check that separately (see JWTConfig.RequireExpiry).
func (c *Claims) IsValid() bool {
	if c.ExpiresAt == 0 {
		return true
	}
	return time.Now().Unix() < c.ExpiresAt
}

// JWTAuth creates JWT authentication middleware.
//
// A misconfigured secret or an unsupported algorithm yields a middleware that
// rejects every request with a diagnostic rather than nil. Returning nil was
// worse than it looks: the documented usage is to register whatever this
// returns, so callers that never checked for nil put a nil in the chain and
// panicked on the first request.
func JWTAuth(cfg JWTConfig) core.MiddlewareFunc {
	switch {
	case len(cfg.Secret) == 0:
		return misconfiguredJWTMiddleware("jwt: secret is empty")
	case cfg.Algorithm == "":
		cfg.Algorithm = jwtAlgHS256
	case cfg.Algorithm != jwtAlgHS256:
		return misconfiguredJWTMiddleware("jwt: unsupported algorithm " + cfg.Algorithm)
	}
	if cfg.TokenName == "" {
		cfg.TokenName = "token"
	}

	return func(ctx *core.Ctx, next core.Handler) error {
		token := extractToken(ctx, cfg)
		if token == "" {
			return exception.Unauthorized("missing authentication token")
		}

		claims, err := parseToken(token, cfg)