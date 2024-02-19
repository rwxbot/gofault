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