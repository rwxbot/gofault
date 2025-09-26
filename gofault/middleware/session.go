package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gofault/gofault/core"
)

// SessionConfig holds session configuration.
type SessionConfig struct {
	Name     string
	MaxAge   int    // Session max age in seconds (0 = until browser closes)
	Path     string // Cookie path
	Domain   string // Cookie domain
	Secure   bool   // Secure cookie flag
	HTTPOnly bool   // HTTPOnly flag
}

// DefaultSessionConfig returns default session configuration.
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		Name:     "gofault_session",
		MaxAge:   86400, // 24 hours
		Path:     "/",
		HTTPOnly: true,
	}
}

// Session represents a user session.
type Session struct {
	ID      string
	Values  map[string]any
	Created time.Time
	Expires time.Time
}
