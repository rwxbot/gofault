package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	content := `
app:
  name: testapp
  environment: testing
  version: 1.0.0
server:
  host: localhost
  port: 8080
database:
  host: 127.0.0.1
  port: 3306
  name: testdb
  user: root
  password: secret
log:
  level: debug
  format: json
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
