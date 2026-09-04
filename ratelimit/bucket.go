package ratelimit

import (
	"sync"
	"time"
)

type Bucket struct {
	mu          sync.Mutex
	capacity    int
	refillEvery time.Duration
	now         func() time.Time
	tokens      int
	lastRefill  time.Time
}

func NewBucket(capacity int, refillEvery time.Duration, now func() time.Time) *Bucket {
	return &Bucket{
		capacity:    capacity,
		refillEvery: refillEvery,
		now:         now,
		tokens:      capacity,
		lastRefill:  now(),
	}
}

func (b *Bucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()
	if b.tokens == 0 {
		return false
	}
	b.tokens--
	return true
}

func (b *Bucket) refill() {
	current := b.now()
	if b.tokens == b.capacity {
		b.lastRefill = current
		return
	}

	elapsedIntervals := int(current.Sub(b.lastRefill) / b.refillEvery)
	if elapsedIntervals == 0 {
		return
	}

	missingTokens := b.capacity - b.tokens
	if elapsedIntervals >= missingTokens {
		b.tokens = b.capacity
		b.lastRefill = current
		return
	}
	b.tokens += elapsedIntervals
	b.lastRefill = b.lastRefill.Add(time.Duration(elapsedIntervals) * b.refillEvery)
}

func (b *Bucket) RetryAfter() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.refill()
	if b.tokens > 0 {
		return 0
	}
	return b.lastRefill.Add(b.refillEvery).Sub(b.now())
}
