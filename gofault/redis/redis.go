// Package redis provides Redis integration for the gofault framework.
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/middleware"
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
		// The client opened a pool before the probe; without closing it a
		// failed NewClient leaks the connection and its goroutines.
		rdb.Close()
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
	return rdb.Del(ctx, keys...).Err()
}

// Exists checks if a key exists.
func Exists(ctx context.Context, rdb *redis.Client, key string) (bool, error) {
	n, err := rdb.Exists(ctx, key).Result()
	return n > 0, err
}

// Lock represents a distributed lock.
type Lock struct {
	rdb     *redis.Client
	key     string
	value   string
	timeout time.Duration
}

// NewLock creates a new distributed lock.
func NewLock(rdb *redis.Client, key, value string, timeout time.Duration) *Lock {
	return &Lock{rdb: rdb, key: key, value: value, timeout: timeout}
}

// ErrLockNotHeld is returned by Release when the lock is no longer held by
// this instance — it expired, or someone else took it over. The Lua script
// result used to be discarded, so callers doing idempotent cleanup or retry
// logic could not tell "released" from "was already gone".
var ErrLockNotHeld = errors.New("redis lock: not held")

// Acquire attempts to acquire the lock.
//
// The value must be non-empty and unique per holder: the release script only
// deletes when GET matches, and two independent holders that both pass ""
// would each satisfy the other's check and release each other's locks.
func (l *Lock) Acquire(ctx context.Context) (bool, error) {
	if l.value == "" {
		return false, fmt.Errorf("redis lock %q: value must not be empty", l.key)
	}
	ok, err := l.rdb.SetNX(ctx, l.key, l.value, l.timeout).Result()
	return ok, err
}

// Release releases the lock if the value matches. It returns ErrLockNotHeld
// when the key is missing or held under a different value.
func (l *Lock) Release(ctx context.Context) error {
	// Use Lua script to ensure atomic check-and-delete
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`)
	res, err := script.Run(ctx, l.rdb, []string{l.key}, l.value).Int64()
	if err != nil {
		return err
	}
	if res == 0 {
		return ErrLockNotHeld
	}
	return nil
}

type redisShutdown struct {
	rdb *redis.Client
}