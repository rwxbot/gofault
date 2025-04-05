package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/prometheus/client_golang/prometheus"
)

func TestDefaultMetricsConfig(t *testing.T) {
	cfg := DefaultMetricsConfig()
	if !cfg.Enabled {