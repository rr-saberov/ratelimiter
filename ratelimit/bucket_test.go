package ratelimit_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rr-saberov/ratelimiter/ratelimit"
)

func TestBucketAllowsRequestsUpToCapacity(t *testing.T) {
	clock := newFakeClock()
	bucket := ratelimit.NewBucket(3, time.Second, clock.Now)

	for i := range 3 {
		if !bucket.Allow() {
			t.Fatalf("request %d: want allowed, got denied", i+1)
		}
	}
}

func TestBucketDeniesRequestsOnceCapacityIsExhausted(t *testing.T) {
	clock := newFakeClock()
	bucket := ratelimit.NewBucket(2, time.Second, clock.Now)
	bucket.Allow()
	bucket.Allow()

	if bucket.Allow() {
		t.Fatal("third request: want denied, got allowed")
	}
}

func TestBucketRefillsOneTokenPerInterval(t *testing.T) {
	clock := newFakeClock()
	bucket := ratelimit.NewBucket(1, time.Second, clock.Now)
	bucket.Allow()

	clock.Advance(500 * time.Millisecond)
	if bucket.Allow() {
		t.Fatal("after half an interval: want denied, got allowed")
	}

	clock.Advance(500 * time.Millisecond)
	if !bucket.Allow() {
		t.Fatal("after a full interval: want allowed, got denied")
	}
}

func TestBucketDoesNotRefillAboveCapacity(t *testing.T) {
	clock := newFakeClock()
	bucket := ratelimit.NewBucket(2, time.Second, clock.Now)
	bucket.Allow()
	bucket.Allow()

	clock.Advance(time.Minute)
	bucket.Allow()
	bucket.Allow()

	if bucket.Allow() {
		t.Fatal("third request after a long idle period: want denied, got allowed")
	}
}

func TestBucketDoesNotAccrueRefillCreditWhileFull(t *testing.T) {
	clock := newFakeClock()
	bucket := ratelimit.NewBucket(1, time.Second, clock.Now)

	clock.Advance(900 * time.Millisecond)
	bucket.Allow()

	clock.Advance(100 * time.Millisecond)
	if bucket.Allow() {
		t.Fatal("100ms after draining a full bucket: want denied, got allowed")
	}

	clock.Advance(900 * time.Millisecond)
	if !bucket.Allow() {
		t.Fatal("one interval after draining: want allowed, got denied")
	}
}

func TestBucketRetryAfterIsZeroWhileTokensRemain(t *testing.T) {
	clock := newFakeClock()
	bucket := ratelimit.NewBucket(1, time.Second, clock.Now)

	if got := bucket.RetryAfter(); got != 0 {
		t.Fatalf("RetryAfter with a token available: want 0, got %v", got)
	}
}

func TestBucketRetryAfterReportsTimeUntilNextToken(t *testing.T) {
	clock := newFakeClock()
	bucket := ratelimit.NewBucket(1, time.Second, clock.Now)
	bucket.Allow()

	clock.Advance(300 * time.Millisecond)
	if got := bucket.RetryAfter(); got != 700*time.Millisecond {
		t.Fatalf("RetryAfter 300ms after draining: want 700ms, got %v", got)
	}
}

func TestBucketIsSafeForConcurrentUse(t *testing.T) {
	clock := newFakeClock()
	bucket := ratelimit.NewBucket(100, time.Second, clock.Now)

	var allowed atomic.Int64
	var wg sync.WaitGroup
	for range 1000 {
		wg.Go(func() {
			if bucket.Allow() {
				allowed.Add(1)
			}
		})
	}
	wg.Wait()

	if got := allowed.Load(); got != 100 {
		t.Fatalf("allowed requests: want 100, got %d", got)
	}
}
