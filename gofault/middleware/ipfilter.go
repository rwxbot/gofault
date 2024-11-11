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

		// Check block list first
		if isIPBlocked(clientIP, blockNets) {
			ctx.Response.WriteHeader(http.StatusForbidden)
			ctx.Response.Write([]byte(`{"status":"error","message":"IP blocked"}`))
			return nil
		}

		// If allow list is not empty, check it
		if len(allowNets) > 0 && !isIPAllowed(clientIP, allowNets) {
			ctx.Response.WriteHeader(http.StatusForbidden)
			ctx.Response.Write([]byte(`{"status":"error","message":"IP not allowed"}`))
			return nil
		}

		return next(ctx)
	}
}

// parseIPPatterns parses IP addresses and CIDR ranges into net.IPNet structures.
func parseIPPatterns(patterns []string) []*net.IPNet {
	var nets []*net.IPNet
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}
		if strings.Contains(pattern, "/") {
			_, ipNet, err := net.ParseCIDR(pattern)
			if err == nil {
				nets = append(nets, ipNet)
			}
		} else {
			ip := net.ParseIP(pattern)
			if ip != nil {
				// Single IP - create a /32 or /128 mask
				mask := net.CIDRMask(32, 32)
				if ip.To4() == nil {
					mask = net.CIDRMask(128, 128)
				}
				nets = append(nets, &net.IPNet{IP: ip, Mask: mask})
			}
		}
	}
	return nets