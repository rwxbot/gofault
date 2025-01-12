package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gorilla/websocket"
)

// WebSocketConfig holds configuration for the WebSocket middleware.
type WebSocketConfig struct {
	// Enabled enables the WebSocket middleware.
	Enabled bool
	// ReadBufferSize is the buffer size for reading messages.
	ReadBufferSize int
	// WriteBufferSize is the buffer size for writing messages.
	WriteBufferSize int
	// ReadTimeout is the read deadline for the connection.
	ReadTimeout time.Duration
	// WriteTimeout is the write deadline for the connection.
	WriteTimeout time.Duration
	// PingInterval is the interval for sending ping frames.
	PingInterval time.Duration
	// CheckOrigin controls origin validation.
	CheckOrigin func(r *http.Request) bool
}
