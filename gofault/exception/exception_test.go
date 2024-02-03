package exception

import (
	"net/http"
	"net/http/httptest"
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