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