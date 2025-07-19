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
