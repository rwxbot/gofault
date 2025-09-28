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

// SessionStore manages sessions in memory.
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
	config   SessionConfig
}

// NewSessionStore creates a new session store.
func NewSessionStore(cfg SessionConfig) *SessionStore {
	if cfg.Name == "" {
		cfg = DefaultSessionConfig()
	}
	return &SessionStore{
		sessions: make(map[string]*Session),
		config:   cfg,
	}
}

// GetSession retrieves a session by ID.
func (s *SessionStore) GetSession(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil, false
	}

	if !session.Expires.IsZero() && time.Now().After(session.Expires) {
		return nil, false
	}

	return session, true
}