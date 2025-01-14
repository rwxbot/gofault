package middleware

import (
	"net/http"
	"net/url"
	"strings"
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
	// CheckOrigin controls origin validation. Nil (default) applies
	// SameOriginCheck: the browser's Origin must match the request Host.
	// gorilla's own nil default accepts every origin, which exposes
	// authenticated users to Cross-Site WebSocket Hijacking.
	CheckOrigin func(r *http.Request) bool
}

// SameOriginCheck is the default origin policy. gorilla/websocket treats a
// nil CheckOrigin as "accept every Origin header" — the comment that used to
// claim "same origin check" was false, and the resulting default let any web
// page open a WebSocket to an authenticated user's server (CSWSH).
//
// A missing Origin header is accepted: non-browser clients (curl, server-to-
// server, mobile apps) do not send one, and the header is only meaningful for
// browser-driven cross-site requests.
func SameOriginCheck(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
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
		CheckOrigin:     nil, // nil: SameOriginCheck (see WebSocketConfig.CheckOrigin)
	}
}

// effectiveCheckOrigin resolves the origin policy for an upgrader. Passing
// config.CheckOrigin straight through would let nil reach gorilla, whose nil
// default accepts every origin.
func effectiveCheckOrigin(config WebSocketConfig) func(*http.Request) bool {
	if config.CheckOrigin != nil {
		return config.CheckOrigin
	}
	return SameOriginCheck
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