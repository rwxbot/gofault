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
	}
	// A zero MaxAge means the session lasts until the browser is closed, per the
	// field's own documentation, so Expires stays the zero value and the expiry
	// checks skip it. Adding 0 used to produce an expiry of now, which made
	// the session expire immediately instead.
	if s.config.MaxAge > 0 {
		session.Expires = time.Now().Add(time.Duration(s.config.MaxAge) * time.Second)
	}

	s.mu.Lock()
	// Drop expired entries first so the cap cannot be reached entirely by dead
	// sessions, which would evict live ones. This scan is throttled because it
	// is O(n) and runs on the request path.
	if len(s.sessions) >= s.config.MaxSessions {
		if time.Since(s.lastReap) > reapInterval {
			s.reapExpiredLocked(time.Now())
			s.lastReap = time.Now()
		}
	}
	for len(s.sessions) >= s.config.MaxSessions {
		if !s.evictOldestLocked() {
			break
		}
	}
	s.sessions[id] = session
	s.order = append(s.order, id)
	s.mu.Unlock()

	return session, nil
}

// DeleteSession removes a session.
//
// It stays in the order queue, which is harmless: evictOldestLocked skips IDs
// that are no longer in the map.
func (s *SessionStore) DeleteSession(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

// Len reports how many sessions are held.
func (s *SessionStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// reapExpiredLocked drops expired sessions. Requires s.mu held for writing.
func (s *SessionStore) reapExpiredLocked(now time.Time) {
	for id, session := range s.sessions {
		if !session.Expires.IsZero() && now.After(session.Expires) {
			delete(s.sessions, id)
		}
	}
}

// evictOldestLocked drops the oldest session and reports whether it removed one.
// Requires s.mu held for writing.
//
// Stale IDs at the front are discarded as they are found, so the queue does not
// grow without bound as sessions expire.
func (s *SessionStore) evictOldestLocked() bool {
	for len(s.order) > 0 {
		oldest := s.order[0]
		s.order = s.order[1:]

		if _, live := s.sessions[oldest]; live {
			delete(s.sessions, oldest)
			return true
		}
	}
	return false
}

// reap drops expired sessions periodically until Close is called.
func (s *SessionStore) reap() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopped:
			return
		case <-ticker.C:
			s.mu.Lock()
			s.reapExpiredLocked(time.Now())
			s.order = s.order[:0]
			s.mu.Unlock()
		}
	}
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

		// Publish the session so downstream handlers can read it via GetSession.
		ctx.SetLocal("session", session)

		err = next(ctx)

		// Only set the cookie when the request succeeded. On error the response
		// may already carry an error body, and issuing a fresh session would
		// suggest the request was accepted.
		if err != nil {
			return err
		}

		httpCookie := &http.Cookie{
			Name:     store.config.Name,
			Value:    session.ID,
			Path:     store.config.Path,
			Domain:   store.config.Domain,
			MaxAge:   store.config.MaxAge,
			Secure:   store.config.Secure,
			HttpOnly: store.config.HTTPOnly,
		}

		// Add rather than Set: a handler may already have set its own cookies,
		// and Set would replace them.
		ctx.Response.Header().Add("Set-Cookie", formatCookie(httpCookie))

		return nil
	}
}

// formatCookie formats a cookie as a Set-Cookie header value.
func formatCookie(c *http.Cookie) string {
	s := c.Name + "=" + c.Value
	if c.Path != "" {
		s += "; Path=" + c.Path
	}
	if c.Domain != "" {
		s += "; Domain=" + c.Domain
	}
	if c.MaxAge > 0 {
		s += "; Max-Age=" + strconv.Itoa(c.MaxAge)
	}
	if c.Secure {
		s += "; Secure"
	}
	if c.HttpOnly {
		s += "; HttpOnly"
	}
	return s
}

// generateSessionID generates a random session ID.
func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GetSession retrieves the session attached to the request by
// SessionMiddleware. It returns nil when the route is not protected by it.
func GetSession(ctx *core.Ctx) *Session {
	if ctx == nil || ctx.Locals == nil {
		return nil
	}
	if s, ok := ctx.Locals["session"].(*Session); ok {
		return s
	}
	return nil
}

// SetValue stores a value in the session.
func (s *Session) SetValue(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Values[key] = value
}

// GetValue retrieves a value from the session.
func (s *Session) GetValue(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.Values[key]
	return v, ok
}

// DeleteValue removes a value from the session.
func (s *Session) DeleteValue(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Values, key)
}

// Len reports how many values the session holds.
func (s *Session) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Values)
}
