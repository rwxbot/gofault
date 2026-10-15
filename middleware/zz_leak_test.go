package middleware

import (
	"testing"
	"time"
)

// A session asked for after it expired must be dropped, not left behind. This is
// the path the reaper cannot cover, since a reaper only runs periodically.
func TestSessionStore_ExpiredSessionIsDroppedOnRead(t *testing.T) {
	cfg := DefaultSessionConfig()
	cfg.MaxAge = 1
	store := NewSessionStore(cfg)
	defer store.Close()

	s, err := store.CreateSession()
	if err != nil {
		t.Fatal(err)
	}

	// Ask for it while valid.
	if _, ok := store.GetSession(s.ID); !ok {
		t.Fatal("session missing before expiry")
	}
	if got := store.Len(); got != 1 {
		t.Fatalf("Len = %d, want 1", got)
	}

	time.Sleep(1200 * time.Millisecond)

	if _, ok := store.GetSession(s.ID); ok {
		t.Error("an expired session was returned as valid")
	}
	if got := store.Len(); got != 0 {
		t.Errorf("Len = %d after the expired session was read, want 0", got)
	}
}

// The map must not grow past the cap when clients never return a cookie.
func TestSessionStore_CapIsEnforced(t *testing.T) {
	cfg := DefaultSessionConfig()
	cfg.MaxSessions = 50
	store := NewSessionStore(cfg)
	defer store.Close()

	for i := 0; i < 1000; i++ {
		if _, err := store.CreateSession(); err != nil {
			t.Fatalf("CreateSession %d: %v", i, err)
		}
	}

	if got := store.Len(); got > 50 {
		t.Errorf("Len = %d, want at most the 50 cap", got)
	}
}

// Live sessions must not be evicted in favour of dead ones.
func TestSessionStore_CapPrefersDroppingExpired(t *testing.T) {
	cfg := DefaultSessionConfig()
	cfg.MaxAge = 1
	cfg.MaxSessions = 20
	store := NewSessionStore(cfg)
	defer store.Close()

	// Fill the store, then let everything expire.
	for i := 0; i < 20; i++ {
		if _, err := store.CreateSession(); err != nil {
			t.Fatal(err)
		}
	}
	time.Sleep(1200 * time.Millisecond)

	// A new session must not have evicted a live one: everything is expired, so
	// there was nothing live to protect.
	for i := 0; i < 5; i++ {
		if _, err := store.CreateSession(); err != nil {
			t.Fatal(err)
		}
	}
	if got := store.Len(); got != 5 {
		t.Errorf("Len = %d, want 5: expired sessions should have been dropped instead of live ones evicted", got)
	}
}

// Close must stop the reaper, otherwise the goroutine outlives the store.
func TestSessionStore_CloseStopsReaper(t *testing.T) {
	cfg := DefaultSessionConfig()
	store := NewSessionStore(cfg)
	store.Close()

	// A second Close must not panic on the closed channel.
	store.Close()
}

// MaxAge of 0 means no expiry, so nothing may be reaped.
func TestSessionStore_NoExpiryWhenMaxAgeZero(t *testing.T) {
	cfg := DefaultSessionConfig()
	cfg.MaxAge = 0
	store := NewSessionStore(cfg)
	defer store.Close()

	s, err := store.CreateSession()
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(50 * time.Millisecond)
	if _, ok := store.GetSession(s.ID); !ok {
		t.Error("a session with no expiry was dropped")
	}
}
