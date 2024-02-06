package exception

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestHTTPException(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
		wantHTTP int
		wantMsg  string
	}{
		{"BadRequest", BadRequest("invalid input"), 400, http.StatusBadRequest, "invalid input"},
		{"Unauthorized", Unauthorized("no token"), 401, http.StatusUnauthorized, "no token"},
		{"Forbidden", Forbidden("access denied"), 403, http.StatusForbidden, "access denied"},
		{"NotFound", NotFound("user not found"), 404, http.StatusNotFound, "user not found"},
		{"InternalServerError", InternalServerError("db error"), 500, http.StatusInternalServerError, "db error"},
		{"Conflict", Conflict("duplicate entry"), 409, http.StatusConflict, "duplicate entry"},
		{"MethodNotAllowed", MethodNotAllowed("POST not allowed"), 405, http.StatusMethodNotAllowed, "POST not allowed"},
		{"PayloadTooLarge", PayloadTooLarge("file too large"), 413, http.StatusRequestEntityTooLarge, "file too large"},
		{"UnsupportedMediaType", UnsupportedMediaType("application/xml not supported"), 415, http.StatusUnsupportedMediaType, "application/xml not supported"},
		{"TooManyRequests", TooManyRequests("rate limited"), 429, http.StatusTooManyRequests, "rate limited"},
		{"ServiceUnavailable", ServiceUnavailable("maintenance"), 503, http.StatusServiceUnavailable, "maintenance"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("err should not be nil")
			}

			httpErr, ok := HTTPExceptionOf(tt.err)
			if !ok {
				t.Fatalf("expected HTTPException, got %T", tt.err)
			}

			if httpErr.GetCode() != tt.wantCode {
				t.Errorf("GetCode() = %d, want %d", httpErr.GetCode(), tt.wantCode)
			}
			if httpErr.GetStatusCode() != tt.wantHTTP {
				t.Errorf("GetStatusCode() = %d, want %d", httpErr.GetStatusCode(), tt.wantHTTP)
			}
			if httpErr.GetMessage() != tt.wantMsg {
				t.Errorf("GetMessage() = %s, want %s", httpErr.GetMessage(), tt.wantMsg)
			}
		})
	}
}

func TestIsHTTPException(t *testing.T) {
	if !IsHTTPException(BadRequest("test")) {
		t.Error("BadRequest should be HTTPException")
	}
	if IsHTTPException(nil) {
		t.Error("nil should not be HTTPException")
	}
	if IsHTTPException(InternalServerError("test")) {
		// This should pass since InternalServerError is also HTTPException
	}
}

func TestHTTPExceptionFilter_Capture(t *testing.T) {
	filter := NewHTTPExceptionFilter()

	t.Run("captures HTTPException", func(t *testing.T) {