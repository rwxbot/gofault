//go:build integration

package gorm_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"
)

// When the pool is exhausted a caller must get a context/deadline error rather
// than block forever. That is the property that lets an application layer its
// own timeout on top, and it is invisible against SQLite, which has no pool.
func TestPoolExhaustion_FailsWithDeadline(t *testing.T) {
	db := openEngine(t, "mysql")
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}

	sqlDB.SetMaxOpenConns(2)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(0)

	var conns []*sql.Conn
	for i := 0; i < 2; i++ {
		c, err := sqlDB.Conn(context.Background())
		if err != nil {
			t.Fatalf("acquiring connection %d of 2: %v", i, err)
		}
		conns = append(conns, c)
	}
	defer func() {
		for _, c := range conns {
			c.Close()
		}
	}()

	// The pool is full. A third acquire must return, not hang.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	_, err = sqlDB.Conn(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error when the pool is exhausted")
	}
	if elapsed > 3*time.Second {
		t.Errorf("took %v to report exhaustion, want a fast failure", elapsed)
	}
	t.Logf("exhaustion surfaced as: %v after %v", err, elapsed)

	// It must be a deadline error, so a caller can distinguish "pool busy"
	// from "database broken".
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("ctx.Err() = %v, want DeadlineExceeded so callers can tell busy from broken", ctx.Err())
	}
}

// Once a connection is returned the pool must accept work again, so exhaustion
// is transient and not a wedge.
func TestPoolExhaustion_RecoversWhenReleased(t *testing.T) {
	db := openEngine(t, "mysql")
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	held, err := sqlDB.Conn(context.Background())
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := held.Close(); err != nil {
		t.Fatalf("release: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var one int
	if err := sqlDB.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		t.Fatalf("query after release: %v", err)
	}
	if one != 1 {
		t.Errorf("SELECT 1 = %d", one)
	}
}

// A concurrent burst larger than the pool must be served by queuing, not by
// losing requests. This is the shape a real load spike takes.
func TestPoolExhaustion_BurstLargerThanPool(t *testing.T) {
	db := openEngine(t, "mysql")
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)

	const workers = 40
	done := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func(i int) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			var n int
			err := sqlDB.QueryRowContext(ctx, "SELECT "+fmt.Sprint(i)).Scan(&n)
			if err == nil && n != i {
				err = fmt.Errorf("got %d, want %d", n, i)
			}
			done <- err
		}(i)
	}

	for i := 0; i < workers; i++ {
		if err := <-done; err != nil {
			t.Errorf("burst worker %d: %v", i, err)
		}
	}
}

// A request that gives up must not hold its connection, which is the leak that
// turns a slow query into a pool that is permanently exhausted.
func TestPoolExhaustion_AbandonedRequestReleasesConnection(t *testing.T) {
	db := openEngine(t, "mysql")
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	// Abandon a query by cancelling its context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := sqlDB.QueryContext(ctx, "SELECT SLEEP(0).Rows();"); err != nil {
		_ = err
	}

	// The pool must still be usable.
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()

	deadline := time.Now().Add(5 * time.Second)
	for {
		var n int
		err := sqlDB.QueryRowContext(ctx2, "SELECT 1").Scan(&n)
		if err == nil {
			if n != 1 {
				t.Errorf("SELECT 1 = %d", n)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("pool never became usable after an abandoned request: %v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
