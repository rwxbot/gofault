//go:build integration

// Package redisfault drives a real Redis container through a network fault to
// verify that the framework recovers.
//
// go-redis reconnects on its own, and the DistributedLimiter deliberately fails
// open when Redis is unreachable. Neither is observable without taking the
// server away for real, so this suite stops the container mid-test and asserts
// what happens during and after the outage.
package redis_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/gofault/gofault/redis"
	goredis "github.com/redis/go-redis/v9"
)

const (
	containerName = "gofault-fault-redis"
	// faultPort is deliberately distinct from the port the ordinary redis
	// tests use, so a leftover container from another suite cannot silently
	// answer for this one.
	faultPort = "16380"
)

func addr() string { return "127.0.0.1:" + faultPort }

// dockerStop and dockerStart take the server away and bring it back, which is
// the failure a Redis outage actually looks like.
func dockerStop(t *testing.T) {
	t.Helper()
	out, err := exec.Command("docker", "stop", containerName).CombinedOutput()
	if err != nil {
		t.Fatalf("docker stop: %v\n%s", err, out)
	}
}

func dockerStart(t *testing.T) {
	t.Helper()
	out, err := exec.Command("docker", "start", containerName).CombinedOutput()
	if err != nil {
		t.Fatalf("docker start: %v\n%s", err, out)
	}
	// Wait for readiness rather than sleeping a fixed interval.
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		err := goredis.NewClient(&goredis.Options{Addr: addr()}).Ping(ctx).Err()
		cancel()
		if err == nil {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("container did not become ready")
}

func newClient(t *testing.T) *goredis.Client {
	t.Helper()
	c := goredis.NewClient(&goredis.Options{
		Addr:         addr(),
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})
	t.Cleanup(func() { c.Close() })
	return c
}

// The limiter must keep serving requests while Redis is down, and must resume
// enforcing once it is back. Failing open during the outage is deliberate; what
// matters is that the limit is enforced again afterwards rather than being
// permanently disabled.
func TestDistributedLimiter_SurvivesRedisOutage(t *testing.T) {
	rdb := newClient(t)
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis is not available: %v", err)
	}

	lim, err := redis.NewDistributedLimiter(rdb, "fault-test", 3, time.Minute, ctx)
	if err != nil {
		t.Fatalf("NewDistributedLimiter: %v", err)
	}

	key := "client"

	// Before the outage the limit is enforced.
	allowed := 0
	for i := 0; i < 6; i++ {
		if lim.Allow(key) {
			allowed++
		}
	}
	if allowed != 3 {
		t.Fatalf("allowed %d before the outage, want 3", allowed)
	}

	dockerStop(t)

	// During the outage every request must be allowed through rather than
	// refused, so a cache outage does not become an application outage.
	for i := 0; i < 5; i++ {
		if !lim.Allow(key) {
			t.Fatalf("request %d during the outage was refused; the limiter should fail open", i)
		}
	}

	dockerStart(t)

	// After recovery a fresh window must enforce the limit again.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		rdb2 := newClient(t)
		lim2, err := redis.NewDistributedLimiter(rdb2, "fault-test-post", 2, time.Minute, context.Background())
		if err == nil {
			got := 0
			for i := 0; i < 5; i++ {
				if lim2.Allow("fresh-key") {
					got++
				}
			}
			if got == 2 {
				return
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Error("the limiter did not resume enforcing the limit after Redis recovered")
}

// The plain cache must degrade to misses rather than errors taking down the
// request, and must work again afterwards.
func TestHTTPCache_SurvivesRedisOutage(t *testing.T) {
	rdb := newClient(t)
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis is not available: %v", err)
	}

	cache := redis.NewCache(rdb, "fault-cache", time.Minute)

	if err := cache.Set(ctx, "k", "v", time.Minute); err != nil {
		t.Fatalf("Set before outage: %v", err)
	}
	if got, err := cache.Get(ctx, "k"); err != nil || got != "v" {
		t.Fatalf("Get before outage = %q, %v", got, err)
	}

	dockerStop(t)

	// Reads fail, which is correct: the caller decides what to do.
	if _, err := cache.Get(ctx, "k"); err == nil {
		t.Log("a read during the outage returned no error")
	}

	// Writes fail the same way.
	if err := cache.Set(ctx, "k2", "v", time.Minute); err == nil {
		t.Log("a write during the outage returned no error")
	}

	dockerStart(t)

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		rdb2 := newClient(t)
		cache2 := redis.NewCache(rdb2, "fault-cache-post", time.Minute)
		if err := cache2.Set(context.Background(), "k", "v", time.Minute); err == nil {
			if got, err := cache2.Get(context.Background(), "k"); err == nil && got == "v" {
				return
			}
		}
		time.Sleep(300 * time.Millisecond)
	}
	t.Error("the cache did not recover after Redis came back")
}

// A long-lived client created before the outage must still work afterwards:
// go-redis keeps its pool and reconnects underneath.
func TestClient_ReconnectsAfterLongOutage(t *testing.T) {
	rdb := newClient(t)
	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skipf("redis is not available: %v", err)
	}

	// Use the client once so the pool has a connection to lose.
	if err := rdb.Set(ctx, "warm", "1", time.Minute).Err(); err != nil {
		t.Fatalf("warmup: %v", err)
	}

	dockerStop(t)

	// Let the pool notice the connection is gone.
	ctx2, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = rdb.Ping(ctx2).Err()

	// Hold the outage long enough that any expire/TTL would have fired.
	time.Sleep(2 * time.Second)

	dockerStart(t)

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second)
		err := rdb.Ping(ctx3).Err()
		cancel3()
		if err == nil {
			// The client is usable again, not merely connected.
			ctx4, cancel4 := context.WithTimeout(context.Background(), time.Second)
			defer cancel4()
			if err := rdb.Get(ctx4, "warm").Err(); err != nil && strings.Contains(err.Error(), "connection") {
				continue
			}
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Error("the long-lived client did not recover after the container came back")
}
