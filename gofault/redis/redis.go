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