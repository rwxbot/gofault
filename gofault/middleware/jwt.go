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