package redis

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Addr != "localhost:6379" {
		t.Errorf("expected Addr=localhost:6379, got %s", cfg.Addr)
	}
	if cfg.PoolSize != 100 {
		t.Errorf("expected PoolSize=100, got %d", cfg.PoolSize)
	}