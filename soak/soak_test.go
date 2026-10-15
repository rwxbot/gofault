//go:build soak

// Package soak runs the assembled framework under sustained mixed traffic and
// watches for growth.
//
// The unit, integration and end-to-end suites catch behaviour that is wrong on
// the first call. They cannot see something that is right on every call but leaks
// a little each time: a goroutine that never exits, a map entry that is never
// dropped, a buffer that is retained. Those only show up as a slope over time.
//
// The test records memory and goroutine counts at intervals, fits a line to the
// second half, and fails when the slope is materially positive. A leak is a
// line, not a point: the first samples include one-off warm-up, so the trend is
// taken from the tail.
package soak

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofault/gofault/exception"
	"github.com/gofault/gofault/middleware"
	"github.com/gofault/gofault/router"
	"github.com/gofault/gofault/server"
)

// sample is one observation of runtime state.
type sample struct {
	at         time.Duration
	heapBytes  uint64
	goroutines int
}

// Harness builds the same stack the e2e suite uses, so the soak measures the
// assembled framework rather than one package.
type harness struct {
	srv  *server.HTTP
	addr string
}

func start(t *testing.T) *harness {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()

	rtr := router.New()
	rtr.ExceptionFilter(exception.NewHTTPExceptionFilter())
	rtr.Handle(http.MethodGet, "/echo",
		wrap(&echoController{}, "Echo"))
	rtr.Handle(http.MethodGet, "/big",
		wrap(&echoController{payloadSize: 8 * 1024}, "Big"))
	rtr.Handle(http.MethodGet, "/items/:id",
		wrap(&echoController{}, "Item"))
	rtr.Handle(http.MethodGet, "/fail",
		wrap(&echoController{}, "Fail"))
	rtr.Handle(http.MethodGet, "/slow",
		wrap(&slowController{delay: 5 * time.Millisecond}, "Slow"))

	// Middleware that allocates and holds state per request. A leak in any of
	// these shows as a slope.
	rtr.Middleware(middleware.RequestID())
	rtr.Middleware(middleware.RecoveryMiddleware(middleware.DefaultRecoveryConfig()))
	rtr.Middleware(middleware.CompressionMiddleware(middleware.DefaultCompressionConfig()))

	limiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		RequestsPerSecond: 1_000_000,
		BurstSize:         1_000_000,
		MaxKeys:           200_000,
	})
	t.Cleanup(limiter.Close)
	rtr.Middleware(limiter.Middleware())

	sessStore := middleware.NewSessionStore(middleware.DefaultSessionConfig())
	rtr.Middleware(middleware.SessionMiddleware(sessStore))

	srv := server.New(rtr, 0)
	srv.Server().Addr = addr
	srv.Server().Handler = rtr

	go func() { _ = srv.Server().Serve(ln) }()
	t.Cleanup(func() { _ = srv.Server().Close() })

	waitFor(t, addr)
	return &harness{srv: srv, addr: addr}
}

func waitFor(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		c, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			c.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("server at %s never came up", addr)
}

func (h *harness) do(client *http.Client, acceptGzip bool, path string) error {
	req, err := http.NewRequest(http.MethodGet, "http://"+h.addr+path, nil)
	if err != nil {
		return err
	}
	if acceptGzip {
		req.Header.Set("Accept-Encoding", "gzip")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		zr, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("gzip reader for %s: %w", path, err)
		}
		defer zr.Close()
		r = zr
	}
	_, err = io.Copy(io.Discard, r)
	return err
}

// worker runs one traffic mix until told to stop.
func worker(t *testing.T, h *harness, stop chan struct{}, wg *sync.WaitGroup,
	ops, failed *int64, errs chan string) {
	t.Helper()
	defer wg.Done()

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 8,
			DisableCompression:  true,
		},
	}

	// Weighted paths: the common case dominates, with error, large-body, slow
	// and parameterised routes mixed in.
	paths := []struct {
		path string
		gzip bool
	}{
		{"/echo", true}, {"/echo", true}, {"/echo", false},
		{"/big", true},
		{"/items/12345", true},
		{"/fail", true},
		{"/slow", true},
	}

	i := 0
	for {
		select {
		case <-stop:
			return
		default:
		}

		p := paths[i%len(paths)]
		i++
		atomic.AddInt64(ops, 1)

		if err := h.do(client, p.gzip, p.path); err != nil {
			atomic.AddInt64(failed, 1)
			select {
			case errs <- err.Error():
			default:
			}
		}
	}
}

// collect samples runtime state while the workers run.
func collect(t *testing.T, d time.Duration, every time.Duration) []sample {
	t.Helper()

	var out []sample
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		runtime.GC()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		out = append(out, sample{
			at:         time.Since(deadline.Add(-d)),
			heapBytes:  m.HeapAlloc,
			goroutines: runtime.NumGoroutine(),
		})

		time.Sleep(every)
	}
	return out
}

// slopePerMinute fits a least-squares line to the tail of the series and reports
// the change per minute. A leak is a line, so the fit is taken from the second
// half, which excludes cold-start warm-up.
func slopePerMinute(samples []sample, value func(sample) float64) float64 {
	if len(samples) < 4 {
		return 0
	}

	half := samples[len(samples)/2:]
	n := float64(len(half))

	var sumX, sumY, sumXY, sumXX float64
	for _, s := range half {
		x := s.at.Minutes()
		y := value(s)
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	denom := n*sumXX - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

func median(xs []int) int {
	if len(xs) == 0 {
		return 0
	}
	sorted := append([]int(nil), xs...)
	sort.Ints(sorted)
	return sorted[len(sorted)/2]
}

func TestSoak_SustainedMixedTraffic(t *testing.T) {
	if testing.Short() {
		t.Skip("soak tests are opt-in; run with -tags soak and without -short")
	}

	duration := 60 * time.Second
	if env := os.Getenv("GOFAULT_SOAK_DURATION"); env != "" {
		if d, err := time.ParseDuration(env); err == nil && d > 0 {
			duration = d
		}
	}
	workers := 16
	if env := os.Getenv("GOFAULT_SOAK_WORKERS"); env != "" {
		if n, err := strconv.Atoi(env); err == nil && n > 0 {
			workers = n
		}
	}

	h := start(t)

	// Warm-up pass so one-off allocations do not look like a leak.
	var warm sync.WaitGroup
	warmStop := make(chan struct{})
	var warmOps int64
	for i := 0; i < 3; i++ {
		warm.Add(1)
		go worker(t, h, warmStop, &warm, &warmOps, new(int64), make(chan string, 1))
	}
	time.Sleep(300 * time.Millisecond)
	close(warmStop)
	warm.Wait()

	runtime.GC()
	var base runtime.MemStats
	runtime.ReadMemStats(&base)

	// Sampler runs alongside, so growth is measured under load rather than
	// only after it stops.
	samplesCh := make(chan []sample, 1)
	go func() {
		samplesCh <- collect(t, duration, 500*time.Millisecond)
	}()

	var ops, failed int64
	errs := make(chan string, 64)
	stop := make(chan struct{})

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(t, h, stop, &wg, &ops, &failed, errs)
	}

	time.Sleep(duration)
	close(stop)
	wg.Wait()

	samples := <-samplesCh

	// Report what happened before judging it.
	t.Logf("requests: %d ok, %d failed over %v with %d workers",
		atomic.LoadInt64(&ops)-warmOps, atomic.LoadInt64(&failed), duration, workers)
	t.Logf("baseline heap: %.1f MiB, goroutines: %d",
		float64(base.HeapAlloc)/(1<<20), runtime.NumGoroutine())

	if len(samples) > 0 {
		first, last := samples[len(samples)/2], samples[len(samples)-1]
		t.Logf("mid-run heap: %.1f MiB, goroutines: %d",
			float64(first.heapBytes)/(1<<20), first.goroutines)
		t.Logf("end heap: %.1f MiB, goroutines: %d",
			float64(last.heapBytes)/(1<<20), last.goroutines)
	}

	heapSlope := slopePerMinute(samples, func(s sample) float64 {
		return float64(s.heapBytes) / (1 << 20)
	})
	goroutineSlope := slopePerMinute(samples, func(s sample) float64 {
		return float64(s.goroutines)
	})
	t.Logf("slope: heap %.2f MiB/min, goroutines %.1f/min",
		heapSlope, goroutineSlope)

	// Thresholds are deliberately loose. A leak shows up as tens of MiB per
	// minute or a goroutine count climbing without levelling; normal variation
	// from GC timing is well under these.
	const maxHeapSlopeMiBPerMin = 12.0
	const maxGoroutineSlopePerMin = 20.0

	if heapSlope > maxHeapSlopeMiBPerMin {
		t.Errorf("heap grew %.2f MiB/min over the second half; want under %.0f",
			heapSlope, maxHeapSlopeMiBPerMin)
	}
	if goroutineSlope > maxGoroutineSlopePerMin {
		t.Errorf("goroutines grew %.1f/min over the second half; want under %.0f",
			goroutineSlope, maxGoroutineSlopePerMin)
	}

	// Requests must not start failing, otherwise a declining error rate would
	// mask the leak by doing less work.
	if f := atomic.LoadInt64(&failed); f > 0 {
		select {
		case e := <-errs:
			t.Fatalf("%d of %d requests failed; a soak with errors measures nothing: %s",
				f, atomic.LoadInt64(&ops), e)
		default:
			t.Fatalf("%d of %d requests failed", f, atomic.LoadInt64(&ops))
		}
	}
}

func medianHeapMiB(samples []sample) float64 {
	bs := make([]int, 0, len(samples))
	for _, s := range samples {
		bs = append(bs, int(s.heapBytes>>20))
	}
	return float64(median(bs))
}

var _ = medianHeapMiB

// The session store's cap must not degrade throughput. Reaching MaxSessions with
// an O(n) scan on every insert cut the soak's request rate by roughly eight
// times, which is invisible to a leak check that only looks at growth.
func TestSoak_SessionCapDoesNotStrangleThroughput(t *testing.T) {
	if testing.Short() {
		t.Skip("soak tests are opt-in; run with -tags soak and without -short")
	}

	cfg := middleware.DefaultSessionConfig()
	cfg.MaxAge = 0 // no expiry, so the cap is what bounds the map
	cfg.MaxSessions = DefaultSessionsCap
	store := middleware.NewSessionStore(cfg)
	defer store.Close()

	// Enough to pass the cap: the interesting path is insert-while-full.
	const perWorker = 2000
	workers := 8

	start := time.Now()
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				if _, err := store.CreateSession(); err != nil {
					t.Errorf("CreateSession: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	total := workers * perWorker
	perSec := float64(total) / elapsed.Seconds()
	t.Logf("created %d sessions in %v (%.0f/s), store holds %d (cap %d)",
		total, elapsed, perSec, store.Len(), cfg.MaxSessions)

	if got := store.Len(); got > cfg.MaxSessions {
		t.Errorf("store holds %d sessions, want at most %d", got, cfg.MaxSessions)
	}

	// The O(1) eviction path should manage well over 10k sessions a second on
	// any modern machine. The O(n) version managed about 1k.
	const minPerSecond = 10000
	if perSec < minPerSecond {
		t.Errorf("session creation ran at %.0f/s, want at least %d; "+
			"the cap is probably being enforced with a scan", perSec, minPerSecond)
	}
}

// DefaultSessionsCap mirrors middleware.DefaultMaxSessions so the throttle check
// runs at the same scale the framework does.
const DefaultSessionsCap = 10000
