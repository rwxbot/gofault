package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gofault/gofault/core"
)

// IPFilterConfig holds configuration for the IP filter middleware.