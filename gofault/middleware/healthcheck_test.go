package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestDefaultHealthCheckConfig(t *testing.T) {
	cfg := DefaultHealthCheckConfig()

	if cfg.Path != "/health" {