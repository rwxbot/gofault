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

// CreateSession creates a new session.
func (s *SessionStore) CreateSession() (*Session, error) {
	id, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:      id,
		Values:  make(map[string]any),
		Created: time.Now(),
		Expires: time.Now().Add(time.Duration(s.config.MaxAge) * time.Second),
	}

	s.mu.Lock()
	s.sessions[id] = session
	s.mu.Unlock()

	return session, nil
}

// DeleteSession removes a session.
func (s *SessionStore) DeleteSession(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

// SessionMiddleware creates session handling middleware.
func SessionMiddleware(store *SessionStore) core.MiddlewareFunc {
	return func(ctx *core.Ctx, next core.Handler) error {
		cookie, err := ctx.Request.Cookie(store.config.Name)
		var session *Session

		if err == nil && cookie != nil && cookie.Value != "" {
			if s, exists := store.GetSession(cookie.Value); exists {
				session = s
			}
		}

		if session == nil {
			session, err = store.CreateSession()
			if err != nil {
				return err
			}
		}

		err = next(ctx)

		// Set session cookie
		httpCookie := &http.Cookie{
			Name:     store.config.Name,
			Value:    session.ID,
			Path:     store.config.Path,
			Domain:   store.config.Domain,
			MaxAge:   store.config.MaxAge,
			Secure:   store.config.Secure,
			HttpOnly: store.config.HTTPOnly,
		}