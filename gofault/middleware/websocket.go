package middleware

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gorilla/websocket"
)

// WebSocketConfig holds configuration for the WebSocket middleware.
type WebSocketConfig struct {
	// Enabled enables the WebSocket middleware.
	Enabled bool
	// ReadBufferSize is the buffer size for reading messages.
	ReadBufferSize int