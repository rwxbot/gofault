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
	// MaxSessions caps how many sessions are held. A client that never sends a
	// cookie back gets a fresh session per request, so without a cap the map
	// grows for as long as the process serves traffic. 0 uses
	// DefaultMaxSessions.
	MaxSessions int
}

// DefaultMaxSessions bounds the session count so an uncached client cannot grow
// the map faster than expired sessions are dropped.
const DefaultMaxSessions = 10000

// DefaultSessionConfig returns default session configuration.
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		Name:        "gofault_session",
		MaxAge:      86400, // 24 hours
		Path:        "/",
		HTTPOnly:    true,
		MaxSessions: DefaultMaxSessions,
	}
}

// Session represents a user session.
//
// Values is guarded by its own mutex. The store's mutex protects the map of
// sessions, not the map inside one, and two requests presenting the same cookie
// share that session: they would otherwise race on Values, and a concurrent
// read and write on a Go map can panic rather than merely return stale data.
//
// Session is not safe to copy because of that mutex; keep the pointer the store
// hands out.
type Session struct {
	ID      string
	Values  map[string]any
	Created time.Time
	Expires time.Time

	mu sync.RWMutex
}

// SessionStore manages sessions in memory.
//
// An expired session is removed when it is next asked for, and a background
// reaper drops the remainder, because a client that never returns the cookie
// creates one session per request: without both, the map grows for as long as
// the process serves traffic. Call Close when the store is no longer needed.
type SessionStore struct {
	sessions map[string]*Session
	// order holds session IDs in creation order. Reaching the cap evicts from
	// the front, which is O(1) amortised; scanning the map for the oldest was
	// O(n) on every insert once the cap was reached, which cut throughput by
	// roughly eight times under the soak test.
	order  []string
	mu     sync.RWMutex
	config SessionConfig

	stopOnce sync.Once
	stopped  chan struct{}

	// lastReap throttles the O(n) expiry scan.
	lastReap time.Time
}

// reapInterval bounds how often the full expiry scan runs on the request path.
// Between scans the cap is maintained by the O(1) order queue.
const reapInterval = 5 * time.Second

// NewSessionStore creates a new session store.
//
// MaxAge of 0 means the session has no expiry, so nothing is ever reaped; set a
// positive MaxAge when the store is expected to serve unbounded traffic.
func NewSessionStore(cfg SessionConfig) *SessionStore {
	if cfg.Name == "" {
		cfg = DefaultSessionConfig()
	}
	if cfg.MaxSessions <= 0 {
		cfg.MaxSessions = DefaultMaxSessions
	}

	s := &SessionStore{
		sessions: make(map[string]*Session),
		config:   cfg,
		stopped:  make(chan struct{}),
	}
	go s.reap()
	return s
}

// Close stops the reaper. Safe to call more than once.
func (s *SessionStore) Close() {
	s.stopOnce.Do(func() { close(s.stopped) })
}

// GetSession retrieves a session by ID, dropping it if it has expired.
func (s *SessionStore) GetSession(id string) (*Session, bool) {
	s.mu.RLock()
	session, exists := s.sessions[id]
	s.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Expiry is decided without the lock so the delete below can take the write
	// lock. A session that disappears concurrently is handled as a miss.
	if !session.Expires.IsZero() && time.Now().After(session.Expires) {
		s.DeleteSession(id)
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