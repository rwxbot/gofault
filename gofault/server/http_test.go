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
	}
}

func TestGracefulShutdown(t *testing.T) {
	handler := http.NewServeMux()
	handler.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	srv := New(handler, 0) // port 0 = random available port

	// Start server
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Logf("server error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(10 * time.Millisecond)

	// Verify server is running
	if srv.IsStopped() {
		t.Error("server should not be stopped yet")
	}

	// Trigger graceful shutdown
	err := srv.GracefulShutdown(5 * time.Second)
	if err != nil {
		t.Errorf("GracefulShutdown() error = %v", err)
	}

	// Verify server is stopped
	if !srv.IsStopped() {
		t.Error("server should be stopped")
	}
}

func TestGracefulShutdown_Hooks(t *testing.T) {
	handler := http.NewServeMux()
	srv := New(handler, 0)

	hook1Called := false
	hook2Called := false

	srv.RegisterShutdownHook(func() {
		hook1Called = true
	})
	srv.RegisterShutdownHook(func() {
		hook2Called = true
	})

	// Start and immediately shutdown
	go func() {
		srv.Start()
	}()
	time.Sleep(10 * time.Millisecond)

	srv.GracefulShutdown(5 * time.Second)

	if !hook1Called {
		t.Error("hook1 should be called")
	}
	if !hook2Called {
		t.Error("hook2 should be called")
	}
}

func TestGracefulShutdown_Idempotent(t *testing.T) {
	srv := New(nil, 0)

	// First shutdown
	srv.GracefulShutdown(5 * time.Second)

	// Second shutdown should be no-op
	err := srv.GracefulShutdown(5 * time.Second)
	if err != nil {
		t.Errorf("second GracefulShutdown() error = %v", err)
	}
}

func TestStop(t *testing.T) {
	handler := http.NewServeMux()
	srv := New(handler, 0)

	// Start server
	go func() {
		srv.Start()
	}()
	time.Sleep(10 * time.Millisecond)

	// Stop (uses default 30s timeout)
	err := srv.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	if !srv.IsStopped() {
		t.Error("server should be stopped")
	}
}

func TestHTTP_ServerState(t *testing.T) {
	srv := New(nil, 8080)

	if srv.IsStopped() {
		t.Error("new server should not be stopped")
	}
}
