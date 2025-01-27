package middleware

import (
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gorilla/websocket"
)

func TestDefaultWebSocketConfig(t *testing.T) {
	cfg := DefaultWebSocketConfig()
	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.ReadBufferSize != 1024 {
		t.Errorf("expected ReadBufferSize 1024, got %d", cfg.ReadBufferSize)
	}
	if cfg.WriteBufferSize != 1024 {
		t.Errorf("expected WriteBufferSize 1024, got %d", cfg.WriteBufferSize)
	}