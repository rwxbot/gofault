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