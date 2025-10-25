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

	// Retrieve it
	got, exists := store.GetSession(session.ID)
	if !exists {
		t.Fatal("session should exist")
	}
	if got.ID != session.ID {
		t.Errorf("got ID = %s, want %s", got.ID, session.ID)
	}
}

func TestSessionStore_GetSession_NotFound(t *testing.T) {
	store := NewSessionStore(DefaultSessionConfig())

	_, exists := store.GetSession("nonexistent")
	if exists {
		t.Error("nonexistent session should not exist")
	}
}

func TestSessionStore_DeleteSession(t *testing.T) {
	store := NewSessionStore(DefaultSessionConfig())

	session, _ := store.CreateSession()
	store.DeleteSession(session.ID)

	_, exists := store.GetSession(session.ID)
	if exists {
		t.Error("session should be deleted")
	}
}

func TestSession_SetValue(t *testing.T) {
	session := &Session{Values: make(map[string]any)}

	session.SetValue("user_id", "123")
	session.SetValue("name", "John")

	if v, ok := session.GetValue("user_id"); !ok || v != "123" {
		t.Errorf("user_id = %v, want 123", v)
	}
	if v, ok := session.GetValue("name"); !ok || v != "John" {
		t.Errorf("name = %v, want John", v)
	}
}

func TestSession_DeleteValue(t *testing.T) {
	session := &Session{Values: make(map[string]any)}
	session.SetValue("temp", "data")

	session.DeleteValue("temp")

	if _, ok := session.GetValue("temp"); ok {
		t.Error("temp value should be deleted")
	}
}

func TestSession_Expires(t *testing.T) {
	cfg := DefaultSessionConfig()
	cfg.MaxAge = 3600 // 1 hour
	store := NewSessionStore(cfg)

	session, _ := store.CreateSession()

	expectedExpiry := session.Created.Add(time.Duration(cfg.MaxAge) * time.Second)
	if session.Expires.Unix() != expectedExpiry.Unix() {
		t.Errorf("Expires = %v, want %v", session.Expires, expectedExpiry)
	}
}

func TestDefaultSessionConfig(t *testing.T) {
	cfg := DefaultSessionConfig()

	if cfg.Name != "gofault_session" {
		t.Errorf("Name = %s, want gofault_session", cfg.Name)
	}
	if cfg.MaxAge != 86400 {
		t.Errorf("MaxAge = %d, want 86400", cfg.MaxAge)
	}
	if cfg.Path != "/" {
		t.Errorf("Path = %s, want /", cfg.Path)
	}
	if !cfg.HTTPOnly {
		t.Error("HTTPOnly should be true")
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1, err := generateSessionID()
	if err != nil {
		t.Fatalf("generateSessionID() error = %v", err)
	}

	id2, _ := generateSessionID()

	if id1 == id2 {
		t.Error("generated IDs should be unique")
	}
	if len(id1) != 64 {
		t.Errorf("ID length = %d, want 64", len(id1))
	}
}

func TestFormatCookie(t *testing.T) {
	c := &http.Cookie{
		Name:     "session",
		Value:    "abc123",
		Path:     "/",
		MaxAge:   3600,
		Secure:   true,
		HttpOnly: true,
	}

	formatted := formatCookie(c)

	expected := "session=abc123; Path=/; Max-Age=3600; Secure; HttpOnly"
	if formatted != expected {
		t.Errorf("formatCookie() = %s, want %s", formatted, expected)
	}
}

func TestSessionMiddleware_Integration(t *testing.T) {
	store := NewSessionStore(DefaultSessionConfig())
	middleware := SessionMiddleware(store)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	ctx := &core.Ctx{Request: req, Response: rec}

	nextCalled := false
	err := middleware(ctx, func(ctx *core.Ctx) error {
		nextCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Error("next handler should be called")
	}
}
