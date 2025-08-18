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

func TestJWTAuth_TokenFromQueryParam(t *testing.T) {
	secret := []byte("test-secret")
	cfg := DefaultJWTConfig(secret)
	middleware := JWTAuth(cfg)

	claims := &Claims{
		Subject:   "user123",
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}
	token, _ := GenerateToken(claims, secret, "HS256")

	req := httptest.NewRequest("GET", "/test?token="+url.QueryEscape(token), nil)
	ctx := &core.Ctx{Request: req}

	err := middleware(ctx, func(ctx *core.Ctx) error { return nil })
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClaims_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt int64
		want      bool
	}{
		{"no expiry", 0, true},
		{"future expiry", time.Now().Add(time.Hour).Unix(), true},
		{"past expiry", time.Now().Add(-time.Hour).Unix(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Claims{ExpiresAt: tt.expiresAt}
			if got := c.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	secret := []byte("test-secret")
	claims := &Claims{
		Subject:   "user123",
		UserID:    "1",
		Roles:     []string{"admin", "user"},
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Issuer:    "test",
	}

	token, err := GenerateToken(claims, secret, "HS256")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Fatal("token should not be empty")
	}

	parts := 0
	for _, c := range token {
		if c == '.' {
			parts++
		}
	}
	if parts != 2 {
		t.Errorf("token should have 3 parts, got %d parts", parts+1)
	}
}

func TestExtractToken_BearerWithSpaces(t *testing.T) {
	cfg := DefaultJWTConfig([]byte("secret"))

	// Test various Bearer formats
	tests := []struct {
		auth   string
		expect string
	}{
		{"Bearer token123", "token123"},
		{"bearer token456", "token456"},
		{"BEARER token789", "token789"},
		{"Basic token", ""}, // Wrong scheme
		{"token", ""},       // No scheme
		{"", ""},            // Empty
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", "/test", nil)
		if tt.auth != "" {
			req.Header.Set("Authorization", tt.auth)
		}
		ctx := &core.Ctx{Request: req}

		got := extractToken(ctx, cfg)
		if got != tt.expect {
			t.Errorf("extractToken(%q) = %q, want %q", tt.auth, got, tt.expect)
		}
	}
}
