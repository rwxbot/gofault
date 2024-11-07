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
}

// DefaultIPFilterConfig returns a default IP filter configuration.
func DefaultIPFilterConfig() IPFilterConfig {
	return IPFilterConfig{
		Enabled: true,
		Mode:    "block",
	}
}

// IPFilterMiddleware creates a middleware that filters requests by IP address.
func IPFilterMiddleware(config IPFilterConfig) core.MiddlewareFunc {
	if !config.Enabled {
		return nil
	}

	allowNets := parseIPPatterns(config.Allow)
	blockNets := parseIPPatterns(config.Block)

	return func(ctx *core.Ctx, next core.Handler) error {
		clientIP := getClientIP(ctx)