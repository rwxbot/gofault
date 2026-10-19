package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultIPFilterConfig(t *testing.T) {
	config := DefaultIPFilterConfig()
	if !config.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if config.Mode != "block" {
		t.Errorf("Expected Mode 'block', got '%s'", config.Mode)
	}
}

func TestIPFilterMiddleware_Disabled(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Enabled = false

	handler := IPFilterMiddleware(config)
	if handler != nil {
		t.Error("Expected nil handler when disabled")
	}
}

func TestIPFilterMiddleware_NoRestrictions(t *testing.T) {
	middleware := IPFilterMiddleware(DefaultIPFilterConfig())

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte(`{"status":"ok"}`))
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	err := middleware(ctx, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestIPFilterMiddleware_BlockList(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Block = []string{"192.168.1.100"}

	middleware := IPFilterMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)
	ctx.Request.RemoteAddr = "192.168.1.100:12345"

	middleware(ctx, next)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestIPFilterMiddleware_AllowList(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Allow = []string{"10.0.0.0/8"}

	middleware := IPFilterMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		ctx.Response.Write([]byte(`{"status":"ok"}`))
		return nil
	}

	// IP in allow range
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)
	ctx.Request.RemoteAddr = "10.1.2.3:12345"

	err := middleware(ctx, next)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	// IP not in allow range
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	ctx2 := core.NewCtx(w2, req2)
	ctx2.Request.RemoteAddr = "192.168.1.1:12345"

	middleware(ctx2, func(c *core.Ctx) error {
		c.Response.WriteHeader(http.StatusOK)
		return nil
	})

	if w2.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w2.Code)
	}
}

func TestParseIPPatterns(t *testing.T) {
	patterns := []string{"192.168.1.1", "10.0.0.0/8", "172.16.0.0/12"}
	nets := parseIPPatterns(patterns)

	if len(nets) != 3 {
		t.Errorf("Expected 3 networks, got %d", len(nets))
	}
}

// The critical finding: without a trusted-proxy gate, any client can send
// "X-Forwarded-For: 127.0.0.1" and walk through an allow-list that permits
// localhost. Proxy headers must be ignored unless the connection itself
// comes from a configured trusted proxy.
func TestIPFilterMiddleware_XFFSpoofingIgnoredByDefault(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Allow = []string{"127.0.0.1"}

	middleware := IPFilterMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	}

	// httptest.NewRequest uses RemoteAddr 192.0.2.1:1234 (TEST-NET-1).
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusForbidden {
		t.Errorf("spoofed XFF must not bypass the allow-list: got %d, want %d", w.Code, http.StatusForbidden)
	}
}

func TestIPFilterMiddleware_XFFHonoredFromTrustedProxy(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Allow = []string{"127.0.0.1"}
	config.TrustedProxies = []string{"192.0.2.0/24"}

	middleware := IPFilterMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	}

	// RemoteAddr 192.0.2.1 is inside the trusted proxy range, so the XFF
	// header set by that proxy is honored.
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("X-Forwarded-For", "127.0.0.1")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusOK {
		t.Errorf("XFF from a trusted proxy should be honored: got %d, want %d", w.Code, http.StatusOK)
	}
}

// A trusted proxy APPENDS its hop to X-Forwarded-For rather than replacing
// it, so a client that forged "X-Forwarded-For: 127.0.0.1" arrives as
// "127.0.0.1, <real client>". Taking the leftmost entry — the old behaviour —
// let the forged value past the allow-list. The chain must be walked right
// to left and the first untrusted hop wins.
func TestIPFilterMiddleware_XFFForgedLeftmostIgnored(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Allow = []string{"127.0.0.1"}
	config.TrustedProxies = []string{"192.0.2.0/24"}

	middleware := IPFilterMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	}

	// The proxy at 192.0.2.1 appended the real client 203.0.113.7 after the
	// client's forged 127.0.0.1.
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.RemoteAddr = "192.0.2.1:5555"
	req.Header.Set("X-Forwarded-For", "127.0.0.1, 203.0.113.7")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusForbidden {
		t.Errorf("forged leftmost XFF must not bypass the allow-list: got %d, want %d", w.Code, http.StatusForbidden)
	}
}

// Behind two trusted proxies the walk skips trusted hops and stops at the
// first untrusted one.
func TestIPFilterMiddleware_XFFSkipsTrustedHops(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Allow = []string{"203.0.113.9"}
	config.TrustedProxies = []string{"192.0.2.0/24", "198.51.100.0/24"}

	middleware := IPFilterMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:5555"
	req.Header.Set("X-Forwarded-For", "203.0.113.9, 198.51.100.2")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusOK {
		t.Errorf("client behind a trusted second hop should resolve to 203.0.113.9: got %d, want %d", w.Code, http.StatusOK)
	}
}

// A malformed hop breaks the chain: entries further left are unverifiable,
// so the walk stops rather than trusting across the garbage.
func TestIPFilterMiddleware_XFFMalformedHopStopsWalk(t *testing.T) {
	config := DefaultIPFilterConfig()
	config.Allow = []string{"203.0.113.9"}
	config.TrustedProxies = []string{"192.0.2.0/24"}

	middleware := IPFilterMiddleware(config)

	next := func(ctx *core.Ctx) error {
		ctx.Response.WriteHeader(http.StatusOK)
		return nil
	}

	// Rightmost hop is garbage: the walk stops immediately and the
	// connection address (a trusted proxy) is used, which is not allowed.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.0.2.1:5555"
	req.Header.Set("X-Forwarded-For", "203.0.113.9, not-an-ip")
	w := httptest.NewRecorder()
	ctx := core.NewCtx(w, req)

	middleware(ctx, next)

	if w.Code != http.StatusForbidden {
		t.Errorf("garbage rightmost hop must stop the walk (fall back to the proxy address): got %d, want %d", w.Code, http.StatusForbidden)
	}
}
