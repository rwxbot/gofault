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

// DefaultWebSocketConfig returns a default WebSocket configuration.
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		Enabled:         true,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    60 * time.Second,
		PingInterval:    30 * time.Second,
		CheckOrigin:     nil, // default: same origin check
	}
}

// WebSocketHandler defines the interface for WebSocket message handlers.
type WebSocketHandler interface {
	// HandleMessage is called when a text or binary message is received.
	HandleMessage(ctx *core.Ctx, conn *websocket.Conn, messageType int, data []byte) error
	// HandleConnect is called when a new WebSocket connection is established.
	HandleConnect(ctx *core.Ctx, conn *websocket.Conn) error
	// HandleDisconnect is called when the WebSocket connection is closed.
	HandleDisconnect(ctx *core.Ctx, conn *websocket.Conn)
}

// WebSocketHandlerFunc is an adapter that allows a plain function to satisfy WebSocketHandler.
// Nil fields are treated as no-ops.
type WebSocketHandlerFunc struct {
	// OnMessage is called for each WebSocket message. Required.
	OnMessage func(ctx *core.Ctx, conn *websocket.Conn, messageType int, data []byte) error
	// OnConnect is called when the connection is established. Optional.
	OnConnect func(ctx *core.Ctx, conn *websocket.Conn) error
	// OnDisconnect is called when the connection is closed. Optional.
	OnDisconnect func(ctx *core.Ctx, conn *websocket.Conn)
}

// HandleMessage implements WebSocketHandler.
func (h WebSocketHandlerFunc) HandleMessage(ctx *core.Ctx, conn *websocket.Conn, messageType int, data []byte) error {
	if h.OnMessage == nil {
		return nil
	}
	return h.OnMessage(ctx, conn, messageType, data)
}

// HandleConnect implements WebSocketHandler.
func (h WebSocketHandlerFunc) HandleConnect(ctx *core.Ctx, conn *websocket.Conn) error {