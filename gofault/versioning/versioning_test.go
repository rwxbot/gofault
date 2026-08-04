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

	if cfg.Strategy != StrategyPathPrefix {
		t.Errorf("expected StrategyPathPrefix, got %d", cfg.Strategy)
	}

	if cfg.PathPrefix != "/v2" {
		t.Errorf("expected path prefix '/v2', got %s", cfg.PathPrefix)
	}

	if cfg.DefaultVersion != 1 {
		t.Errorf("expected default version 1, got %d", cfg.DefaultVersion)
	}
}

func TestQueryConfig(t *testing.T) {
	cfg := QueryConfig("api_version", 1)

	if cfg.Strategy != StrategyQuery {
		t.Errorf("expected StrategyQuery, got %d", cfg.Strategy)
	}

	if cfg.QueryParam != "api_version" {
		t.Errorf("expected query param 'api_version', got %s", cfg.QueryParam)
	}

	if cfg.DefaultVersion != 1 {
		t.Errorf("expected default version 1, got %d", cfg.DefaultVersion)
	}
}

func TestVersionStatus_String(t *testing.T) {
	tests := []struct {
		status VersionStatus
		want   string
	}{
		{VersionStatusActive, "active"},
		{VersionStatusDeprecated, "deprecated"},
		{VersionStatusEOL, "eol"},
		{VersionStatus(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("VersionStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestVersionSet_Add(t *testing.T) {
	vs := &VersionSet{}
	vs.Add(VersionInfo{Version: 1, Status: VersionStatusActive})