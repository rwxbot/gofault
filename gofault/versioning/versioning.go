// Package versioning provides API versioning support for HTTP routers.
package versioning

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofault/gofault/core"
)

// VersionStatus indicates the lifecycle state of an API version.
type VersionStatus int

const (
	// VersionStatusActive is a currently supported version.
	VersionStatusActive VersionStatus = iota
	// VersionStatusDeprecated indicates a version pending retirement.
	VersionStatusDeprecated
	// VersionStatusEOL signals end-of-life and should not be used.
	VersionStatusEOL
)

func (s VersionStatus) String() string {
	switch s {
	case VersionStatusActive:
		return "active"
	case VersionStatusDeprecated:
		return "deprecated"
	case VersionStatusEOL:
		return "eol"
	default:
		return "unknown"
	}
}

// Strategy defines how the version is extracted from a request.
type Strategy int

const (
	// StrategyHeader extracts version from a header (e.g., Accept).
	StrategyHeader Strategy = iota
	// StrategyPathPrefix extracts version from URL path prefix.
	StrategyPathPrefix
	// StrategyQuery extracts version from query parameter.
	StrategyQuery
)

// Config holds versioning middleware configuration.
type Config struct {
	// Strategy for version extraction.
	Strategy Strategy
	// Header is used with StrategyHeader: header name to read.
	Header string
	// HeaderVersionFormat is used with StrategyHeader: printf format (e.g., "v%d" or "version=%s").
	HeaderVersionFormat string
	// HeaderVersionRE is used with StrategyHeader: regex to extract version after matching headerVersionFormat.
	HeaderVersionRE *regexp.Regexp
	// PathPrefix is used with StrategyPathPrefix: URL path prefix (e.g., "/v1").
	PathPrefix string
	// QueryParam is used with StrategyQuery: query parameter name (e.g., "version").
	QueryParam string
	// DefaultVersion returned when no version is detected.
	DefaultVersion int
	// DefaultStatus for versions without explicit status.
	DefaultStatus VersionStatus
}

// HeaderConfig creates a header-based versioning config.
// headerVersionFormat uses %d for numeric version, %s for string version.
// Example: "application/vnd.api+json;version=v%d" extracts "1" from "v1".
func HeaderConfig(header, headerVersionFormat string, headerVersionRE *regexp.Regexp, defaultVersion int) Config {
	return Config{
		Strategy:            StrategyHeader,
		Header:              header,
		HeaderVersionFormat: headerVersionFormat,
		HeaderVersionRE:     headerVersionRE,
		DefaultVersion:      defaultVersion,
		DefaultStatus:       VersionStatusActive,
	}
}

// PathPrefixConfig creates a path-prefix based versioning config.