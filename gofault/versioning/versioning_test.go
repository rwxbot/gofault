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

	info, ok := vs.Get(1)
	if !ok {
		t.Fatal("expected version 1 to exist")
	}
	if info.Status != VersionStatusActive {
		t.Errorf("expected active, got %s", info.Status)
	}
}

func TestVersionSet_Get(t *testing.T) {
	vs := DefaultVersionSet([]VersionInfo{
		{Version: 1, Status: VersionStatusActive},
		{Version: 2, Status: VersionStatusDeprecated, Message: "Use v3"},
	})

	info, ok := vs.Get(1)
	if !ok {
		t.Fatal("expected version 1")
	}
	if info.Status != VersionStatusActive {
		t.Errorf("v1: expected active, got %s", info.Status)
	}

	_, ok = vs.Get(999)
	if ok {
		t.Error("version 999 should not exist")
	}
}

func TestVersionSet_IsSupported(t *testing.T) {
	vs := DefaultVersionSet([]VersionInfo{
		{Version: 1, Status: VersionStatusActive},
		{Version: 2, Status: VersionStatusDeprecated},
		{Version: 3, Status: VersionStatusEOL},
	})

	if !vs.IsSupported(1) {
		t.Error("v1 active should be supported")
	}

	if !vs.IsSupported(2) {
		t.Error("v2 deprecated should still be supported")
	}

	if vs.IsSupported(3) {
		t.Error("v3 EOL should not be supported")
	}

	if vs.IsSupported(999) {
		t.Error("unknown version should not be supported")
	}
}

func TestVersionSet_IsDeprecated(t *testing.T) {
	vs := DefaultVersionSet([]VersionInfo{
		{Version: 1, Status: VersionStatusActive},
		{Version: 2, Status: VersionStatusDeprecated},
	})

	if vs.IsDeprecated(1) {
		t.Error("v1 active should not be deprecated")
	}

	if !vs.IsDeprecated(2) {
		t.Error("v2 deprecated should be deprecated")
	}
}

func TestVersionSet_Deprecate(t *testing.T) {
	vs := DefaultVersionSet([]VersionInfo{{Version: 1, Status: VersionStatusActive}})
	vs.Deprecate(1, "Use v2 instead")

	info, _ := vs.Get(1)
	if info.Status != VersionStatusDeprecated {
		t.Errorf("expected deprecated, got %s", info.Status)
	}
	if info.Message != "Use v2 instead" {
		t.Errorf("expected message 'Use v2 instead', got %q", info.Message)
	}
}

func TestVersionSet_EOL(t *testing.T) {
	vs := DefaultVersionSet([]VersionInfo{{Version: 1, Status: VersionStatusActive}})
	vs.EOL(1, "No longer available")

	info, _ := vs.Get(1)
	if info.Status != VersionStatusEOL {
		t.Errorf("expected eol, got %s", info.Status)
	}

	if vs.IsSupported(1) {
		t.Error("EOL version should not be supported")
	}
}

func TestVersionSet_Deprecate_UnknownVersion(t *testing.T) {
	vs := &VersionSet{}
	vs.Deprecate(99, "does not exist")

	info, ok := vs.Get(99)
	if !ok {
		t.Fatal("version should be created")
	}
	if info.Status != VersionStatusDeprecated {
		t.Errorf("expected deprecated, got %s", info.Status)
	}
}
