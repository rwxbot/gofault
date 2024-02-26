// Package server provides the HTTP server implementation.
package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// HTTP wraps the standard library HTTP server.
type HTTP struct {
	server  *http.Server
	hooks   []func()
	mu      sync.RWMutex
	stopped bool
}

// New creates a new HTTP server that will listen on the given port.
func New(handler http.Handler, port int) *HTTP {
	return &HTTP{
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handler,
		},
	}
}

// NewWithConfig creates a new HTTP server with custom configuration.
func NewWithConfig(handler http.Handler, addr string) *HTTP {
	return &HTTP{
		server: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

// RegisterShutdownHook registers a function to be called during graceful shutdown.
func (s *HTTP) RegisterShutdownHook(hook func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hooks = append(s.hooks, hook)
}

// Start launches the server and blocks.
func (s *HTTP) Start() error {
	return s.server.ListenAndServe()
}

// StartWithGracefulShutdown starts the server and waits for shutdown signals.
func (s *HTTP) StartWithGracefulShutdown(timeout time.Duration) error {
	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return s.GracefulShutdown(timeout)
}

// GracefulShutdown shuts down the server gracefully with a timeout.
func (s *HTTP) GracefulShutdown(timeout time.Duration) error {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return nil
	}
	s.stopped = true
	s.mu.Unlock()

	// Execute shutdown hooks
	s.mu.RLock()
	hooks := s.hooks
	s.mu.RUnlock()
	for _, hook := range hooks {
		hook()
	}

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Shutdown server
	return s.server.Shutdown(ctx)
}

// Stop gracefully shuts down the server with a default 30 second timeout.
func (s *HTTP) Stop() error {
	return s.GracefulShutdown(30 * time.Second)
}

// IsStopped returns whether the server has been stopped.
func (s *HTTP) IsStopped() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stopped
}
