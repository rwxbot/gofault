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
	if h.OnConnect == nil {
		return nil
	}
	return h.OnConnect(ctx, conn)
}

// HandleDisconnect implements WebSocketHandler.
func (h WebSocketHandlerFunc) HandleDisconnect(ctx *core.Ctx, conn *websocket.Conn) {
	if h.OnDisconnect != nil {
		h.OnDisconnect(ctx, conn)
	}
}

// WebSocketMiddleware creates a middleware that upgrades HTTP to WebSocket.
func WebSocketMiddleware(config WebSocketConfig, handler WebSocketHandler) core.MiddlewareFunc {
	if !config.Enabled {
		return nil
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
		CheckOrigin:     effectiveCheckOrigin(config),
	}

	return func(ctx *core.Ctx, next core.Handler) error {
		conn, err := upgrader.Upgrade(ctx.Response, ctx.Request, nil)
		if err != nil {
			return err
		}
		defer conn.Close()

		// Set deadlines
		conn.SetReadDeadline(time.Now().Add(config.ReadTimeout))
		conn.SetWriteDeadline(time.Now().Add(config.WriteTimeout))

		// Start ping handler
		stopCh := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(config.PingInterval)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(config.WriteTimeout)); err != nil {
						return
					}
				case <-stopCh:
					return