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