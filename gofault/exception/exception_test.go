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