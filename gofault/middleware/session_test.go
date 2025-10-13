package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofault/gofault/core"
)

func TestNewSessionStore(t *testing.T) {
	cfg := DefaultSessionConfig()
	store := NewSessionStore(cfg)

	if store == nil {
		t.Fatal("store should not be nil")
	}
	if store.sessions == nil {
		t.Error("sessions map should be initialized")
	}
}

func TestSessionStore_CreateSession(t *testing.T) {
	store := NewSessionStore(DefaultSessionConfig())

	session, err := store.CreateSession()
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	if session.ID == "" {
		t.Error("session ID should not be empty")
	}
	if len(session.ID) != 64 { // 32 bytes = 64 hex chars
		t.Errorf("session ID length = %d, want 64", len(session.ID))
	}
	if session.Values == nil {
		t.Error("session Values should be initialized")
	}
	if session.Created.IsZero() {
		t.Error("session Created should be set")
	}
}

func TestSessionStore_GetSession(t *testing.T) {
	store := NewSessionStore(DefaultSessionConfig())

	// Create a session
	session, _ := store.CreateSession()
