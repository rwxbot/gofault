package middleware

import (
	"net"
	"sync"
	"time"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

// RateLimiterConfig holds rate limiter configuration.
type RateLimiterConfig struct {