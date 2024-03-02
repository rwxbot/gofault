package server

import (
	"net/http"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	handler := http.NewServeMux()
	srv := New(handler, 8080)

	if srv == nil {
		t.Fatal("server should not be nil")
	}
	if srv.server == nil {
		t.Error("http.Server should not be nil")
	}
	if srv.server.Addr != ":8080" {
		t.Errorf("Addr = %s, want :8080", srv.server.Addr)
	}
}

func TestNewWithConfig(t *testing.T) {
	handler := http.NewServeMux()
	srv := NewWithConfig(handler, "localhost:9090")

	if srv.server.Addr != "localhost:9090" {
		t.Errorf("Addr = %s, want localhost:9090", srv.server.Addr)
	}
}

func TestRegisterShutdownHook(t *testing.T) {
	srv := New(nil, 8080)

	srv.RegisterShutdownHook(func() {
		// hook registration test
	})

	// We can't easily test hook execution without a full shutdown,
	// but we can verify registration doesn't panic
	srv.RegisterShutdownHook(func() {})
	if len(srv.hooks) != 2 {
		t.Errorf("hooks count = %d, want 2", len(srv.hooks))