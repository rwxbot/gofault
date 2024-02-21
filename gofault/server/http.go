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