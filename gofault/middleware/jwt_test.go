package middleware

import (
	"net/http/httptest"
	"net/url"
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