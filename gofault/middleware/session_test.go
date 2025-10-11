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