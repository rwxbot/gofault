package redis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/gofault/gofault/core"
	"github.com/gofault/gofault/exception"
)

// The distributed limiter must enforce the limit across what are, as far as
// Redis is concerned, separate callers -- which is what multiple replicas look
// like. A per-process limiter would let each one through.
func TestDistributedLimiter_EnforcesSharedLimit(t *testing.T) {
	rdb := newTestClient(t)
	ctx := context.Background()

	const limit = 5
	lim, err := NewDistributedLimiter(rdb, "shared-limit-test", limit, time.Minute, ctx)
	if err != nil {
		t.Fatalf("NewDistributedLimiter: %v", err)
	}

	key := "client-a"
	allowed := 0
	for i := 0; i < 20; i++ {
		if lim.Allow(key) {
			allowed++
		}
	}

	if allowed != limit {
		t.Errorf("allowed %d requests, want %d", allowed, limit)
	}
}

// Each key gets its own budget.
func TestDistributedLimiter_PerKey(t *testing.T) {
	rdb := newTestClient(t)
	lim, err := NewDistributedLimiter(rdb, "per-key-test", 2, time.Minute, context.Background())
	if err != nil {
		t.Fatalf("NewDistributedLimiter: %v", err)
	}

	for i := 0; i < 5; i++ {
		if lim.Allow("client-a") {
			// first two allowed
		}
	}
	// client-b must still have its own budget.
	if !lim.Allow("client-b") {
		t.Error("a second key should have its own budget")
	}
	if !lim.Allow("client-b") {
		t.Error("a second key should have its own budget (second)")
	}
	if lim.Allow("client-b") {
		t.Error("the second key should be exhausted after its limit")
	}
}

// Two limiters pointing at the same Redis are the multi-replica case: they must
// share the count, not each grant a full quota.
func TestDistributedLimiter_ReplicasShareCount(t *testing.T) {
	rdb := newTestClient(t)
	ctx := context.Background()

	// Same key prefix and window, so they are two replicas of one limiter.
	a, err := NewDistributedLimiter(rdb, "replica-test", 6, time.Minute, ctx)
	if err != nil {
		t.Fatalf("replica a: %v", err)
	}
	b, err := NewDistributedLimiter(rdb, "replica-test", 6, time.Minute, ctx)
	if err != nil {
		t.Fatalf("replica b: %v", err)
	}

	key := "shared"
	allowed := func() int {
		n := 0
		if a.Allow(key) {
			n++
		}
		if b.Allow(key) {
			n++
		}
		return n
	}

	total := 0
	for round := 0; round < 10; round++ {
		total += allowed()
	}

	if total != 6 {
		t.Errorf("two replicas allowed %d requests together, want 6", total)
	}
}

// The window must expire, otherwise the key is blocked forever.
func TestDistributedLimiter_WindowExpires(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping window expiry, which needs real time to pass")
	}

	rdb := newTestClient(t)
	lim, err := NewDistributedLimiter(rdb, "expiry-test", 1, 200*time.Millisecond, context.Background())
	if err != nil {
		t.Fatalf("NewDistributedLimiter: %v", err)
	}

	key := "k"
	if !lim.Allow(key) {
		t.Fatal("first request should be allowed")
	}
	if lim.Allow(key) {
		t.Fatal("second request within the window should be refused")
	}

	wait := 400 * time.Millisecond
	time.Sleep(wait)

	if !lim.Allow(key) {
		t.Error("the request should be allowed again once the window rolled")
	}
}

// A Redis outage must fail open: the limiter must not take the application down
// with it.
func TestDistributedLimiter_FailsOpenOnUnreachableRedis(t *testing.T) {
	// A client pointed at a closed port.
	rdb := goredis.NewClient(&goredis.Options{Addr: "127.0.0.1:1"})
	defer rdb.Close()

	lim, err := NewDistributedLimiter(rdb, "unreachable", 1, time.Minute, context.Background())
	if err != nil {
		t.Fatalf("NewDistributedLimiter: %v", err)
	}

	for i := 0; i < 5; i++ {
		if !lim.Allow("k") {
			t.Fatalf("request %d was refused while Redis is down; the limiter should fail open", i)
		}
	}
}

// Count reports the observed state, which is how a test or a metrics endpoint
// can verify the limiter without guessing.
func TestDistributedLimiter_Count(t *testing.T) {
	rdb := newTestClient(t)
	lim, err := NewDistributedLimiter(rdb, "count-test", 10, time.Minute, context.Background())
	if err != nil {
		t.Fatalf("NewDistributedLimiter: %v", err)
	}

	now := time.Now()
	for i := 0; i < 3; i++ {
		lim.Allow("k")
	}

	n, err := lim.Count("k", now)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 3 {
		t.Errorf("Count = %d, want 3", n)
	}
}

func TestNewDistributedLimiter_RejectsBadConfig(t *testing.T) {
	rdb := newTestClient(t)

	if _, err := NewDistributedLimiter(rdb, "x", 0, time.Minute, nil); err == nil {
		t.Error("limit 0 should be rejected")
	}
	if _, err := NewDistributedLimiter(rdb, "x", 5, 0, nil); err == nil {
		t.Error("a zero window should be rejected")
	}
	if _, err := NewDistributedLimiter(nil, "x", 5, time.Minute, nil); err == nil {
		t.Error("a nil client should be rejected")
	}
}

// An empty key must not panic or produce an unusable Redis key.
func TestDistributedLimiter_EmptyKey(t *testing.T) {
	rdb := newTestClient(t)
	lim, err := NewDistributedLimiter(rdb, "empty-key-test", 2, time.Minute, context.Background())
	if err != nil {
		t.Fatalf("NewDistributedLimiter: %v", err)
	}

	if !lim.Allow("") {
		t.Error("an empty key should still be tracked, not refused")
	}
}

// The limiter must work through the middleware, which is how it is actually
// wired. This is the part a unit test of the limiter alone cannot prove.
func TestDistributedLimiter_ThroughMiddleware(t *testing.T) {
	rdb := newTestClient(t)
	lim, err := NewDistributedLimiter(rdb, "mw-test", 3, time.Minute, context.Background())
	if err != nil {
		t.Fatalf("NewDistributedLimiter: %v", err)
	}

	mw := lim.Middleware()
	keyFn := func(ctx *core.Ctx) string { return "fixed-key" }
	_ = keyFn

	var allowed int
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/x", nil)
		w := httptest.NewRecorder()
		ctx := core.NewCtx(w, req)

		err := mw(ctx, func(ctx *core.Ctx) error {
			w.WriteHeader(http.StatusOK)
			return nil
		})

		if err != nil {
			// middleware returned an exception
			if he, ok := exception.HTTPExceptionOf(err); ok && he.GetStatusCode() == http.StatusTooManyRequests {
				continue
			}
			t.Fatalf("unexpected error: %v", err)
		}
		if w.Code == http.StatusOK {
			allowed++
		}
	}

	if allowed != 3 {
		t.Errorf("middleware allowed %d, want 3", allowed)
	}
}
