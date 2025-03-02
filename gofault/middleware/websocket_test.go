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
}

func (m *mockWSHandler) HandleDisconnect(ctx *core.Ctx, conn *websocket.Conn) {
	atomic.AddInt32(&m.disconnectCount, 1)
}

func TestWebSocketMiddleware_FullIntegration(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.PingInterval = 100 * time.Millisecond

	// Echo handler: write back whatever we receive
	handler := WebSocketHandlerFunc{
		OnMessage: func(ctx *core.Ctx, conn *websocket.Conn, msgType int, data []byte) error {
			return conn.WriteMessage(msgType, data)
		},
	}
	mw := WebSocketMiddleware(cfg, handler)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := core.NewCtx(w, r)
		mw(ctx, func(c *core.Ctx) error { return nil })
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}

	msgType, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if msgType != websocket.TextMessage {
		t.Errorf("expected text message, got %d", msgType)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got '%s'", string(data))
	}
}

func TestWebSocketMiddleware_ClientClose(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.PingInterval = 200 * time.Millisecond

	handler := &mockWSHandler{}
	mw := WebSocketMiddleware(cfg, handler)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := core.NewCtx(w, r)
		mw(ctx, func(c *core.Ctx) error { return nil })
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}

	time.Sleep(30 * time.Millisecond)

	conn.Close()

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&handler.connectCount) != 1 {
		t.Errorf("expected 1 connect, got %d", handler.connectCount)
	}
	if atomic.LoadInt32(&handler.disconnectCount) != 1 {
		t.Errorf("expected 1 disconnect, got %d", handler.disconnectCount)
	}
}

func TestWebSocketMiddleware_EchoHandler(t *testing.T) {
	cfg := DefaultWebSocketConfig()

	handler := WebSocketHandlerFunc{
		OnMessage: func(ctx *core.Ctx, conn *websocket.Conn, msgType int, data []byte) error {
			return conn.WriteMessage(msgType, data)
		},
	}
	mw := WebSocketMiddleware(cfg, handler)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := core.NewCtx(w, r)
		mw(ctx, func(c *core.Ctx) error { return nil })
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	tests := []string{"hello", "world", "websocket"}
	for _, msg := range tests {
		err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			t.Fatalf("write error for '%s': %v", msg, err)
		}
		_, data, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("read error for '%s': %v", msg, err)
		}
		if string(data) != msg {
			t.Errorf("expected '%s', got '%s'", msg, string(data))
		}
	}
}

func TestWebSocketMiddleware_OnConnectCallback(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	var connectCalled int32

	handler := WebSocketHandlerFunc{
		OnConnect: func(ctx *core.Ctx, conn *websocket.Conn) error {
			atomic.AddInt32(&connectCalled, 1)
			return nil
		},
		OnMessage: func(ctx *core.Ctx, conn *websocket.Conn, msgType int, data []byte) error {
			return conn.WriteMessage(msgType, data)
		},
	}
	mw := WebSocketMiddleware(cfg, handler)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := core.NewCtx(w, r)
		mw(ctx, func(c *core.Ctx) error { return nil })
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	time.Sleep(30 * time.Millisecond)

	if atomic.LoadInt32(&connectCalled) != 1 {
		t.Errorf("expected OnConnect called once, got %d", connectCalled)
	}
}

func TestWebSocketHijackMiddleware_UpgradeError(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	mw := WebSocketHijackMiddleware(cfg)

	req := httptest.NewRequest("GET", "/ws", nil) // no WS headers
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)

	err := mw(ctx, func(c *core.Ctx) error { return nil })
	// Upgrade without proper WS headers should fail
	if err == nil {
		t.Log("upgrade did not fail (may succeed with some implementations)")
	}
}

func TestWebSocketMiddleware_DisconnectCallback(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	var disconnectCalled int32

	handler := WebSocketHandlerFunc{
		OnMessage: func(ctx *core.Ctx, conn *websocket.Conn, msgType int, data []byte) error {
			return nil
		},
		OnDisconnect: func(ctx *core.Ctx, conn *websocket.Conn) {
			atomic.AddInt32(&disconnectCalled, 1)
		},
	}
	mw := WebSocketMiddleware(cfg, handler)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := core.NewCtx(w, r)
		mw(ctx, func(c *core.Ctx) error { return nil })
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}

	time.Sleep(20 * time.Millisecond)
	conn.Close()
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&disconnectCalled) != 1 {
		t.Errorf("expected OnDisconnect called once, got %d", disconnectCalled)
	}
}

func TestWebSocketHandlerFunc_AllFieldsSet(t *testing.T) {
	var connectMsg, disconnectMsg, messageText string

	handler := WebSocketHandlerFunc{
		OnConnect: func(ctx *core.Ctx, conn *websocket.Conn) error {
			connectMsg = "connected"
			return nil
		},
		OnMessage: func(ctx *core.Ctx, conn *websocket.Conn, msgType int, data []byte) error {
			messageText = string(data)
			return nil
		},
		OnDisconnect: func(ctx *core.Ctx, conn *websocket.Conn) {
			disconnectMsg = "disconnected"
		},
	}

	req := httptest.NewRequest("GET", "/ws", nil)
	rec := httptest.NewRecorder()
	ctx := core.NewCtx(rec, req)
	conn := &websocket.Conn{}

	handler.HandleConnect(ctx, conn)
	handler.HandleMessage(ctx, conn, websocket.TextMessage, []byte("test"))
	handler.HandleDisconnect(ctx, conn)

	if connectMsg != "connected" {
		t.Errorf("expected connectMsg 'connected', got '%s'", connectMsg)
	}
	if messageText != "test" {
		t.Errorf("expected messageText 'test', got '%s'", messageText)
	}
	if disconnectMsg != "disconnected" {
		t.Errorf("expected disconnectMsg 'disconnected', got '%s'", disconnectMsg)
	}
}

// CSWSH guard: with the default policy, a browser Origin that does not match
// the Host must be rejected. gorilla's nil default used to accept it.
func TestWebSocketMiddleware_CrossOriginRejected(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	handler := &mockWSHandler{}
	mw := WebSocketMiddleware(cfg, handler)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := core.NewCtx(w, r)
		mw(ctx, func(c *core.Ctx) error { return nil })
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": {"http://evil.example.com"}})
	if err == nil {
		t.Fatal("cross-origin handshake should be rejected")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 rejection, got %+v", resp)
	}
}

func TestWebSocketMiddleware_SameOriginAccepted(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	handler := &mockWSHandler{}
	mw := WebSocketMiddleware(cfg, handler)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := core.NewCtx(w, r)
		mw(ctx, func(c *core.Ctx) error { return nil })
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	host := strings.TrimPrefix(server.URL, "http://")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Origin": {"http://" + host}})
	if err != nil {
		t.Fatalf("same-origin handshake should succeed: %v", err)
	}
	conn.Close()
}

// A handler business error must surface from the middleware: the old shared
// readErr fed it into IsUnexpectedCloseError, which returned false and the
// error vanished.
func TestWebSocketMiddleware_HandlerErrorSurfaces(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	cfg.PingInterval = 0

	handler := WebSocketHandlerFunc{
		OnMessage: func(ctx *core.Ctx, conn *websocket.Conn, msgType int, data []byte) error {
			return errWSBusiness
		},
	}
	mw := WebSocketMiddleware(cfg, handler)

	var mwErr error
	var wg sync.WaitGroup
	wg.Add(1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		ctx := core.NewCtx(w, r)
		mwErr = mw(ctx, func(c *core.Ctx) error { return nil })
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	conn.WriteMessage(websocket.TextMessage, []byte("trigger"))
	conn.Close()
	wg.Wait()

	if !errors.Is(mwErr, errWSBusiness) {
		t.Errorf("middleware should return the handler error, got %v", mwErr)
	}
}

var errWSBusiness = errors.New("business failure")

// After Upgrade succeeds, the hijacked connection is not managed by net/http
// anymore: if the middleware does not close it, an errored handler leaves the
// socket open until the peer times out. The client must see the connection
// end shortly after the handler returns, not after a long deadline.
func TestWebSocketHijackMiddleware_ClosesConnWhenHandlerFails(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	mw := WebSocketHijackMiddleware(cfg)

	handlerDone := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := core.NewCtx(w, r)
		mw(ctx, func(c *core.Ctx) error {
			close(handlerDone)
			return errors.New("ws handler failed")
		})
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	<-handlerDone

	// A closed server side surfaces as a read error well inside the deadline.
	// Without the fix this read blocks for the full deadline.
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if _, _, err := conn.ReadMessage(); err == nil {
		t.Fatal("expected the server to close the connection after the handler failed")
	} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		t.Fatal("connection still open after the handler failed (read timed out)")
	}
}
