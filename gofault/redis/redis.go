// Package redis provides Redis integration for the gofault framework.
package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/redis/go-redis/v9"
)

// Config holds Redis configuration.
type Config struct {
	// Addr is the Redis address (host:port).
	Addr string
	// Password for Redis AUTH (empty = no auth).
	Password string
	// DB is the database number (0-15).
	DB int
	// PoolSize sets the maximum number of open connections.
	PoolSize int
	// MinIdleConns sets the minimum idle connections.
	MinIdleConns int
	// DialTimeout sets the connection timeout.
	DialTimeout int // in seconds
	// ReadTimeout sets the read timeout.
	ReadTimeout int // in seconds
	// WriteTimeout sets the write timeout.
	WriteTimeout int // in seconds
	// ConnMaxLifetime sets the maximum lifetime of a connection.
	ConnMaxLifetime int // in seconds, 0 = no limit
}

// DefaultConfig returns a default Redis configuration.
func DefaultConfig() Config {
	return Config{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		PoolSize:     100,
		MinIdleConns: 10,
		DialTimeout:  5,
		ReadTimeout:  3,
		WriteTimeout: 3,
	}
}

// Client is the Redis module.
type Client struct {
	core.Module
	RDB    *redis.Client
	Config Config
}

var (
	clientInstance *Client
	clientOnce     sync.Once
)

// NewClient creates a new Redis client module.
func NewClient(name string, config Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  time.Duration(config.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
		ConnMaxLifetime: func() time.Duration {
			if config.ConnMaxLifetime == 0 {
				return 0
			}
			return time.Duration(config.ConnMaxLifetime) * time.Second
		}(),
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	c := &Client{
		Module: *core.NewModule(name),
		RDB:    rdb,
		Config: config,
	}

	c.RegisterOnShutdown(&redisShutdown{rdb})

	return c, nil
}

// MustNewClient creates a new Redis client and panics on error.
func MustNewClient(name string, config Config) *Client {
	c, err := NewClient(name, config)
	if err != nil {
		panic(err)
	}
	return c
}

// GetClient returns the underlying redis.Client.
func (c *Client) GetClient() *redis.Client {
	return c.RDB
}

// Set stores a key-value pair with an optional expiration.
func Set(ctx context.Context, rdb *redis.Client, key string, value interface{}, expiration time.Duration) error {
	return rdb.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key.
func Get(ctx context.Context, rdb *redis.Client, key string) (string, error) {
	return rdb.Get(ctx, key).Result()
}

// Delete removes a key.
func Delete(ctx context.Context, rdb *redis.Client, keys ...string) error {