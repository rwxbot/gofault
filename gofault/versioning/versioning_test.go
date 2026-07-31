package versioning

import (
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gofault/gofault/core"
)

func TestHeaderConfig(t *testing.T) {
	re := regexp.MustCompile(`version=v(\d+)`)
	cfg := HeaderConfig("Accept", "application/vnd.api+json;version=v%d", re, 1)
