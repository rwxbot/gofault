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
