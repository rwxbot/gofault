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