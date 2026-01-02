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