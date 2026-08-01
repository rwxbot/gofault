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

	if cfg.Strategy != StrategyHeader {
		t.Errorf("expected StrategyHeader, got %d", cfg.Strategy)
	}

	if cfg.Header != "Accept" {
		t.Errorf("expected header 'Accept', got %s", cfg.Header)
	}

	if cfg.DefaultVersion != 1 {
		t.Errorf("expected default version 1, got %d", cfg.DefaultVersion)
	}
}

func TestPathPrefixConfig(t *testing.T) {
	cfg := PathPrefixConfig("/v2", 1)