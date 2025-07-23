package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

// JWTConfig holds JWT authentication configuration.
type JWTConfig struct {
	Secret    []byte
	Algorithm string // HS256 only for simplicity
	TokenName string
}

// DefaultJWTConfig returns default JWT configuration.
func DefaultJWTConfig(secret []byte) JWTConfig {
	return JWTConfig{
		Secret:    secret,
		Algorithm: "HS256",
		TokenName: "token",
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

// IsValid checks if claims are valid.
func (c *Claims) IsValid() bool {
	if c.ExpiresAt == 0 {
		return true
	}
	return time.Now().Unix() < c.ExpiresAt
}

// JWTAuth creates JWT authentication middleware.
func JWTAuth(cfg JWTConfig) core.MiddlewareFunc {
	if cfg.TokenName == "" {
		cfg.TokenName = "token"
	}

	return func(ctx *core.Ctx, next core.Handler) error {
		token := extractToken(ctx, cfg)
		if token == "" {
			return exception.Unauthorized("missing authentication token")
		}

		claims, err := parseToken(token, cfg)
		if err != nil {
			return exception.Unauthorized("invalid token: " + err.Error())
		}

		if !claims.IsValid() {
			return exception.Unauthorized("token expired")
		}

		// Hand the verified claims to downstream handlers. Without this the
		// middleware authenticates the caller but nothing downstream can read
		// who the caller is.
		ctx.SetLocal("claims", claims)

		return next(ctx)
	}
}

// GetClaims returns the claims verified by JWTAuth, or nil when the route is not
// protected by it.
func GetClaims(ctx *core.Ctx) *Claims {
	if claims, ok := ctx.Locals["claims"].(*Claims); ok {
		return claims
	}
	return nil
}
