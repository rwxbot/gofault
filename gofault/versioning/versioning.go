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