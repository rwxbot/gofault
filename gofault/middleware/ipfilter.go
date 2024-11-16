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
//
// Even behind a trusted proxy the FIRST entry of X-Forwarded-For is not the
// client: proxies append their hop rather than replacing the header, so a
// client that sent "X-Forwarded-For: 127.0.0.1" arrives with
// "127.0.0.1, <client>". The chain is therefore scanned right to left — the
// rightmost entry was written by the proxy we are connected to — and the
// first entry not itself a trusted proxy is taken as the client. Entries
// further left are older hops that earlier proxies recorded and may be
// forged; they are only reachable once every closer hop is trusted.
func getClientIP(ctx *core.Ctx, trustedNets []*net.IPNet) net.IP {
	remote := remoteIP(ctx)
	if remote == nil {
		return nil
	}

	if !isTrusted(remote, trustedNets) {
		return remote
	}

	xff := ctx.Request.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			ip := net.ParseIP(strings.TrimSpace(parts[i]))
			if ip == nil {
				// A malformed hop breaks the chain: everything further left
				// is unverifiable, so stop rather than trust across it.
				break
			}
			if !isTrusted(ip, trustedNets) {
				return ip
			}
		}
		// Every entry was a trusted proxy (or the header was garbage): the
		// closest untrusted hop is the connection itself.
		return remote
	}

	// X-Real-IP is a single value set by the proxy in front of us. It is only
	// consulted when the connection is trusted, and unlike XFF there is no
	// chain to walk — the proxy that set it is the hop we accepted.
	xri := ctx.Request.Header.Get("X-Real-IP")
	if xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip
		}
	}

	return remote
}

// isTrusted reports whether ip falls inside one of the trusted proxy nets.
func isTrusted(ip net.IP, trustedNets []*net.IPNet) bool {
	for _, ipNet := range trustedNets {
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
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
