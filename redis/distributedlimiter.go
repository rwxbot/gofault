package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/middleware"
)

// DistributedLimiter enforces a rate limit that is shared across every replica
// of an application.
//
// middleware.RateLimiter keeps buckets in the process, so with N replicas the
// effective limit is N times the configured one. This limiter keeps the count in
// Redis, which every replica sees, so the limit is the limit.
//
// It uses a fixed window rather than the token bucket, because a window is a
// single atomic Redis operation. The trade-off is a burst at the boundary: a
// client can use its full quota just before the window rolls and again just
// after, doubling the steady-state rate for one interval.
type DistributedLimiter struct {
	rdb       *goredis.Client
	prefix    string
	limit     int64
	window    time.Duration
	keyPrefix string
	ctx       context.Context
}

// NewDistributedLimiter builds a limiter allowing limit requests per window.
//
// window must be positive, limit must be at least 1, and ctx bounds every
// Redis operation.
func NewDistributedLimiter(rdb *goredis.Client, keyPrefix string, limit int, window time.Duration, ctx context.Context) (*DistributedLimiter, error) {
	if rdb == nil {
		return nil, errors.New("redis: DistributedLimiter needs a client")
	}
	if limit < 1 {
		return nil, fmt.Errorf("redis: limit must be at least 1, got %d", limit)
	}
	if window <= 0 {
		return nil, fmt.Errorf("redis: window must be positive, got %v", window)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	return &DistributedLimiter{
		rdb:       rdb,
		prefix:    "ratelimit",
		limit:     int64(limit),
		window:    window,
		keyPrefix: keyPrefix,
		ctx:       ctx,
	}, nil
}

// allowScript increments the window counter and reports whether the caller is
// over the limit. Both steps are one operation, so concurrent requests across
// replicas cannot interleave a read between the increment and the expiry.
var allowScript = goredis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
if current > tonumber(ARGV[1]) then
    return 0
end
return 1
`)

// Allow implements middleware.Limiter.
//
// A Redis failure allows the request: the alternative would turn a cache outage
// into a full outage of the application behind it.
func (l *DistributedLimiter) Allow(key string) bool {
	if key == "" {
		key = "unknown"
	}

	res, err := allowScript.Run(l.ctx, l.rdb,
		[]string{l.windowKey(key, time.Now())},
		l.limit, l.window.Milliseconds()).Result()
	if err != nil {
		return true
	}

	allowed, ok := res.(int64)
	if !ok {
		return true
	}
	return allowed == 1
}

// Close exists to satisfy middleware.Limiter. There is nothing to release: the
// caller owns the Redis client so it can close it when the whole application
// shuts down.
func (l *DistributedLimiter) Close() {}

func (l *DistributedLimiter) fullKey(k string) string {
	if l.keyPrefix == "" {
		return k
	}
	return l.keyPrefix + ":" + k
}

// Middleware is RateLimitMiddleware over this limiter, so it can be wired
// wherever the in-process one is.
func (l *DistributedLimiter) Middleware() core.MiddlewareFunc {
	return middleware.RateLimitMiddleware(l, nil)
}

// windowKey builds the Redis key for one key in one window.
//
// The caller key is hashed rather than interpolated: it comes from a request
// header, so it can be long and can contain colons, which would otherwise let a
// crafted value collide with another key's namespace or inflate the keyspace.
func (l *DistributedLimiter) windowKey(key string, now time.Time) string {
	sum := sha256.Sum256([]byte(key))
	window := now.UnixMilli() / l.window.Milliseconds()
	return l.fullKey(fmt.Sprintf("%s:%s:%d", l.prefix, hex.EncodeToString(sum[:16]), window))
}

// Count reports the current count for key in the window containing now, so
// callers and tests can observe the state without guessing.
func (l *DistributedLimiter) Count(key string, now time.Time) (int64, error) {
	return l.rdb.Get(l.ctx, l.windowKey(key, now)).Int64()
}

// Compile-time check that the limiter satisfies the middleware contract.
var _ middleware.Limiter = (*DistributedLimiter)(nil)
