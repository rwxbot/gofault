package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/prometheus/client_golang/prometheus"
)

// MetricsConfig holds configuration for the metrics middleware.