package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestRequestID_GeneratesNewID(t *testing.T) {
	middleware := RequestID()

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()