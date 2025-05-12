package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gofault/gofault/core"
)

// CacheConfig holds configuration for the cache middleware.
type CacheConfig struct {
	// TTL is the time-to-live for cache entries.