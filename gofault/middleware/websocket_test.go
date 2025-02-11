package middleware

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gorilla/websocket"
)

func TestDefaultWebSocketConfig(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.ReadBufferSize != 1024 {
		t.Errorf("expected ReadBufferSize 1024, got %d", cfg.ReadBufferSize)
	}
	if cfg.WriteBufferSize != 1024 {
		t.Errorf("expected WriteBufferSize 1024, got %d", cfg.WriteBufferSize)
	}
	if cfg.ReadTimeout != 60*time.Second {
		t.Errorf("expected ReadTimeout 60s, got %v", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 60*time.Second {
		t.Errorf("expected WriteTimeout 60s, got %v", cfg.WriteTimeout)
	}
	if cfg.PingInterval != 30*time.Second {
		t.Errorf("expected PingInterval 30s, got %v", cfg.PingInterval)
	}
}

func TestWebSocketMiddleware_Disabled(t *testing.T) {
	cfg := WebSocketConfig{Enabled: false}
	mw := WebSocketMiddleware(cfg, nil)
	if mw != nil {
		t.Error("expected nil middleware when disabled")
	}
}

func TestWebSocketHandlerFunc_HandleMessage(t *testing.T) {
	var called int32
	var receivedMsg []byte
	var receivedType int

	handler := WebSocketHandlerFunc{
		OnMessage: func(ctx *core.Ctx, conn *websocket.Conn, msgType int, data []byte) error {
			atomic.AddInt32(&called, 1)
			receivedType = msgType
			receivedMsg = data
			return nil
		},
	}

	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)
	conn := &websocket.Conn{}

	err := handler.HandleMessage(ctx, conn, websocket.TextMessage, []byte("hello"))
	if err != nil {
		t.Fatalf("HandleMessage returned error: %v", err)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Error("handler was not called")
	}
	if receivedType != websocket.TextMessage {
		t.Errorf("expected text message type, got %d", receivedType)
	}
	if string(receivedMsg) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(receivedMsg))
	}
}

func TestWebSocketHandlerFunc_NilOnMessage(t *testing.T) {
	handler := WebSocketHandlerFunc{}
	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)
	conn := &websocket.Conn{}

	// Nil OnMessage should not panic and should return nil
	err := handler.HandleMessage(ctx, conn, websocket.TextMessage, []byte("hello"))
	if err != nil {
		t.Errorf("HandleMessage with nil OnMessage returned error: %v", err)
	}
}

func TestWebSocketHandlerFunc_NilOnConnect(t *testing.T) {
	handler := WebSocketHandlerFunc{}
	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)
	conn := &websocket.Conn{}

	err := handler.HandleConnect(ctx, conn)
	if err != nil {
		t.Errorf("HandleConnect with nil OnConnect returned error: %v", err)
	}
}

func TestWebSocketHandlerFunc_NilOnDisconnect(t *testing.T) {
	handler := WebSocketHandlerFunc{}
	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)
	conn := &websocket.Conn{}

	// Should not panic
	handler.HandleDisconnect(ctx, conn)
}

func TestWebSocketHandlerFunc_ImplementsInterface(t *testing.T) {
	var _ WebSocketHandler = WebSocketHandlerFunc{}
}

func TestWebSocketHijackMiddleware_Disabled(t *testing.T) {
	cfg := WebSocketConfig{Enabled: false}
	mw := WebSocketHijackMiddleware(cfg)
	if mw != nil {
		t.Error("expected nil middleware when disabled")
	}
}

func TestWebSocketMiddleware_CheckOriginNil(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.CheckOrigin = nil

	mw := WebSocketMiddleware(cfg, WebSocketHandlerFunc{})
	if mw == nil {
		t.Error("middleware should not be nil")
	}
}

func TestWebSocketHijackMiddleware_StoresConnection(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	mw := WebSocketHijackMiddleware(cfg)

	// Create a test server with proper WebSocket upgrade
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := core.NewCtx(w, r)
		var capturedConn *websocket.Conn
		var capturedUpgrader *websocket.Upgrader
		mw(ctx, func(c *core.Ctx) error {
			capturedConn, _ = c.Locals["ws_conn"].(*websocket.Conn)
			capturedUpgrader, _ = c.Locals["ws_upgrader"].(*websocket.Upgrader)
			if capturedConn == nil {
				t.Error("connection not stored in ctx.Locals")
			}
			if capturedUpgrader == nil {
				t.Error("upgrader not stored in ctx.Locals")
			}
			return nil
		})
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()
}

// mockWSHandler is a test implementation of WebSocketHandler.
type mockWSHandler struct {
	connectCount    int32
	disconnectCount int32
	messageCount    int32
	mu              sync.Mutex
	messages        []string
}

func (m *mockWSHandler) HandleConnect(ctx *core.Ctx, conn *websocket.Conn) error {
	atomic.AddInt32(&m.connectCount, 1)
	return nil
}

func (m *mockWSHandler) HandleMessage(ctx *core.Ctx, conn *websocket.Conn, messageType int, data []byte) error {
	atomic.AddInt32(&m.messageCount, 1)
	m.mu.Lock()
	m.messages = append(m.messages, string(data))
	m.mu.Unlock()
	return nil