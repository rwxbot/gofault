package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultRequestLoggerConfig(t *testing.T) {
	config := DefaultRequestLoggerConfig()