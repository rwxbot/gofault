package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

func TestJWTAuth_MissingToken(t *testing.T) {
	cfg := DefaultJWTConfig([]byte("secret"))
	middleware := JWTAuth(cfg)

	req := httptest.NewRequest("GET", "/test", nil)
	ctx := &core.Ctx{Request: req}

	err := middleware(ctx, func(ctx *core.Ctx) error { return nil })
	if err == nil {
		t.Fatal("expected unauthorized error")
	}

	httpErr, ok := exception.HTTPExceptionOf(err)
	if !ok {
		t.Fatalf("expected HTTPException, got %T", err)
	}
	if httpErr.GetStatusCode() != 401 {
		t.Errorf("status = %d, want 401", httpErr.GetStatusCode())
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	secret := []byte("test-secret")
	cfg := DefaultJWTConfig(secret)
	middleware := JWTAuth(cfg)

	claims := &Claims{
		Subject:   "user123",
		UserID:    "1",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	token, err := GenerateToken(claims, secret, "HS256")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	ctx := &core.Ctx{Request: req}

	nextCalled := false
	err = middleware(ctx, func(ctx *core.Ctx) error {
		nextCalled = true
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !nextCalled {
		t.Error("next handler should be called")
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	secret := []byte("test-secret")
	cfg := DefaultJWTConfig(secret)
	middleware := JWTAuth(cfg)

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(-time.Hour).Unix(), // Expired
	}
	token, _ := GenerateToken(claims, secret, "HS256")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	ctx := &core.Ctx{Request: req}

	err := middleware(ctx, func(ctx *core.Ctx) error { return nil })
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestJWTAuth_InvalidSignature(t *testing.T) {
	secret := []byte("test-secret")
	cfg := DefaultJWTConfig(secret)
	middleware := JWTAuth(cfg)

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	token, _ := GenerateToken(claims, []byte("wrong-secret"), "HS256")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	ctx := &core.Ctx{Request: req}

	err := middleware(ctx, func(ctx *core.Ctx) error { return nil })
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
}

func TestJWTAuth_BearerCaseInsensitive(t *testing.T) {
	secret := []byte("test-secret")
	cfg := DefaultJWTConfig(secret)
	middleware := JWTAuth(cfg)

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	token, _ := GenerateToken(claims, secret, "HS256")

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "BEARER "+token)
	ctx := &core.Ctx{Request: req}

	err := middleware(ctx, func(ctx *core.Ctx) error { return nil })
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// Query-string tokens are off by default because they leak into access logs,
// Referer headers and browser history.
func TestJWTAuth_TokenFromQueryParamDisabledByDefault(t *testing.T) {
	secret := []byte("test-secret")
	middleware := JWTAuth(DefaultJWTConfig(secret))

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	token, _ := GenerateToken(claims, secret, "HS256")

	req := httptest.NewRequest("GET", "/test?token="+url.QueryEscape(token), nil)
	ctx := &core.Ctx{Request: req}

	err := middleware(ctx, func(ctx *core.Ctx) error { return nil })
	if err == nil {
		t.Fatal("expected rejection of query token, got nil")
	}
}

func TestJWTAuth_TokenFromQueryParamAllowedWhenOptedIn(t *testing.T) {
	secret := []byte("test-secret")
	cfg := DefaultJWTConfig(secret)
	cfg.AllowQueryToken = true
	middleware := JWTAuth(cfg)

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	token, _ := GenerateToken(claims, secret, "HS256")

	req := httptest.NewRequest("GET", "/test?token="+url.QueryEscape(token), nil)
	ctx := &core.Ctx{Request: req}