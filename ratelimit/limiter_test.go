package ratelimit_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rr-saberov/ratelimiter/ratelimit"
)

func TestLimiterLimitsEachKeyIndependently(t *testing.T) {
	clock := newFakeClock()
	limiter := ratelimit.NewLimiter(1, time.Second, clock.Now)

	if !limiter.Allow("alice") {
		t.Fatal("alice's first request: want allowed, got denied")
	}
	if !limiter.Allow("bob") {
		t.Fatal("bob's first request: want allowed, got denied")
	}
	if limiter.Allow("alice") {
		t.Fatal("alice's second request: want denied, got allowed")
	}
}

func TestLimiterTracksRetryAfterPerKey(t *testing.T) {
	clock := newFakeClock()
	limiter := ratelimit.NewLimiter(1, time.Second, clock.Now)
	limiter.Allow("alice")

	if got := limiter.RetryAfter("alice"); got != time.Second {
		t.Fatalf("alice RetryAfter: want 1s, got %v", got)
	}
	if got := limiter.RetryAfter("bob"); got != 0 {
		t.Fatalf("bob RetryAfter: want 0, got %v", got)
	}
}

func TestLimiterIsSafeForConcurrentUse(t *testing.T) {
	clock := newFakeClock()
	limiter := ratelimit.NewLimiter(10, time.Second, clock.Now)
	keys := []string{"alice", "bob", "carol"}

	var allowed atomic.Int64
	var wg sync.WaitGroup
	for i := range 300 {
		wg.Go(func() {
			if limiter.Allow(keys[i%len(keys)]) {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()

	if got := allowed.Load(); got != 30 {
		t.Fatalf("allowed requests across 3 keys: want 30, got %d", got)
	}
}
