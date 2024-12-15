package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofault/gofault/core"
)

// TimeoutConfig holds configuration for the timeout middleware.
type TimeoutConfig struct {
	// Enabled enables the timeout middleware.
	Enabled bool