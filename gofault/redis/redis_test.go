package redis

import (
	"context"
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
	if cfg.MinIdleConns != 10 {
		t.Errorf("expected MinIdleConns=10, got %d", cfg.MinIdleConns)
	}
}

func TestNewClient_Invalid(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Addr = "localhost:9999" // no redis running here

	_, err := NewClient("test", cfg)
	if err == nil {
		t.Error("expected error for unreachable Redis")
	}
}

func TestConfig_ZeroConnMaxLifetime(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ConnMaxLifetime = 0 // means no limit
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Addr,
		ConnMaxLifetime: func() time.Duration {
			if cfg.ConnMaxLifetime == 0 {
				return 0
			}
			return time.Duration(cfg.ConnMaxLifetime) * time.Second
		}(),
	})
	if rdb == nil {
		t.Fatal("rdb is nil")
	}
	rdb.Close()
}

// redisAddr resolves the address of the Redis instance used by the integration
// tests below. It honours GOFAULT_REDIS_ADDR so CI can point the tests at a
// provisioned service, and otherwise falls back to the conventional local port.
func redisAddr() string {
	if addr := os.Getenv("GOFAULT_REDIS_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:6379"
}

// newTestClient dials Redis and skips the calling test when no server answers.