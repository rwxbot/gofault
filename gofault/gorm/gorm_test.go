package gorm

import (
	"fmt"
	"testing"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxOpenConns != 100 {
		t.Errorf("expected MaxOpenConns=100, got %d", cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns != 10 {
		t.Errorf("expected MaxIdleConns=10, got %d", cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime != 3600 {
		t.Errorf("expected ConnMaxLifetime=3600, got %d", cfg.ConnMaxLifetime)
	}
	if cfg.LogLevel != logger.Warn {
		t.Errorf("expected LogLevel=Warn, got %v", cfg.LogLevel)
	}
}

func TestNewDatabase_SQLite(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Dialect = DialectSQLite
	cfg.DSN = ":memory:"
	cfg.Silent = true

	db, err := NewDatabase("test-sqlite", cfg)
	if err != nil {
		t.Fatalf("NewDatabase failed: %v", err)
	}