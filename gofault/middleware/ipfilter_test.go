package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultIPFilterConfig(t *testing.T) {
	config := DefaultIPFilterConfig()
	if !config.Enabled {