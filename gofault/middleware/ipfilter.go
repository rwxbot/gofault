package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gofault/gofault/core"
)

// IPFilterConfig holds configuration for the IP filter middleware.
type IPFilterConfig struct {
	// Enabled enables the IP filter middleware.
	Enabled bool
	// Allow is the list of allowed IP addresses or CIDR ranges. If empty, all are allowed except blocked.
	Allow []string
	// Block is the list of blocked IP addresses or CIDR ranges.
	Block []string
	// Mode determines the filter behavior when an IP matches both allow and block lists.
	// "block" (default): blocked IPs take precedence.
	// "allow": allowed IPs take precedence.
	Mode string
	// TrustedProxies lists CIDR ranges or IPs of reverse proxies between the
	// internet and this server. X-Forwarded-For and X-Real-IP are only
	// consulted when the connection's RemoteAddr is one of them. Without
	// this gate any client could send "X-Forwarded-For: 127.0.0.1" and
	// bypass the entire filter. Empty (default): proxy headers are ignored
	// and the TCP remote address is authoritative.
	TrustedProxies []string
}

// DefaultIPFilterConfig returns a default IP filter configuration.
func DefaultIPFilterConfig() IPFilterConfig {
	return IPFilterConfig{
		Enabled: true,
		Mode:    "block",
	}
}

// IPFilterMiddleware creates a middleware that filters requests by IP address.