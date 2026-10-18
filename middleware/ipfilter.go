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
func IPFilterMiddleware(config IPFilterConfig) core.MiddlewareFunc {
	if !config.Enabled {
		return nil
	}

	allowNets := parseIPPatterns(config.Allow)
	blockNets := parseIPPatterns(config.Block)
	trustedNets := parseIPPatterns(config.TrustedProxies)

	return func(ctx *core.Ctx, next core.Handler) error {
		clientIP := getClientIP(ctx, trustedNets)
		blocked := isIPBlocked(clientIP, blockNets)
		allowed := len(allowNets) == 0 || isIPAllowed(clientIP, allowNets)

		// Mode decides which list wins when an IP matches both. It was
		// documented but never consulted, so Mode="allow" silently behaved
		// like "block".
		var deny bool
		switch config.Mode {
		case "allow":
			deny = !allowed
		default: // "block" and any unrecognised value
			deny = blocked || !allowed
		}

		if deny {
			msg := "IP blocked"
			if !blocked {
				msg = "IP not allowed"
			}
			ctx.Response.Header().Set("Content-Type", "application/json")
			ctx.Response.WriteHeader(http.StatusForbidden)
			ctx.Response.Write([]byte(`{"status":"error","message":"` + msg + `"}`))
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
}

// remoteIP returns the authoritative IP of the TCP connection.
func remoteIP(ctx *core.Ctx) net.IP {
	host, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
	if err == nil {
		ip := net.ParseIP(host)
		if ip != nil {
			return ip
		}
	}
	return net.ParseIP(ctx.Request.RemoteAddr)
}

// getClientIP extracts the client IP from the request.
//
// X-Forwarded-For and X-Real-IP are attacker-controlled unless the request
// arrived through a trusted reverse proxy: unconditionally trusting them let
// any client spoof its way past the filter (e.g. "X-Forwarded-For: 127.0.0.1"
// against an allow-list that permits localhost). The headers are therefore
// only consulted when the connection's RemoteAddr is inside TrustedProxies;
// with no trusted proxies configured the TCP address is authoritative.
func getClientIP(ctx *core.Ctx, trustedNets []*net.IPNet) net.IP {
	remote := remoteIP(ctx)
	if remote == nil {
		return nil
	}

	trusted := false
	for _, ipNet := range trustedNets {
		if ipNet.Contains(remote) {
			trusted = true
			break
		}
	}
	if !trusted {
		return remote
	}

	// Check X-Forwarded-For header first
	xff := ctx.Request.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the chain
		parts := strings.Split(xff, ",")
		ip := net.ParseIP(strings.TrimSpace(parts[0]))
		if ip != nil {
			return ip
		}
	}

	// Check X-Real-IP header
	xri := ctx.Request.Header.Get("X-Real-IP")
	if xri != "" {
		ip := net.ParseIP(xri)
		if ip != nil {
			return ip
		}
	}

	return remote
}

// isIPBlocked checks if an IP is in the block list.
func isIPBlocked(ip net.IP, blockList []*net.IPNet) bool {
	for _, ipNet := range blockList {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}

// isIPAllowed checks if an IP is in the allow list.
func isIPAllowed(ip net.IP, allowList []*net.IPNet) bool {
	for _, ipNet := range allowList {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}
