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
