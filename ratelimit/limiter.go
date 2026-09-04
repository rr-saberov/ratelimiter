package ratelimit

import (
	"sync"
	"time"
)

type Limiter struct {
	mu          sync.Mutex
	buckets     map[string]*Bucket
	capacity    int
	refillEvery time.Duration
	now         func() time.Time
}

func NewLimiter(capacity int, refillEvery time.Duration, now func() time.Time) *Limiter {
	return &Limiter{
		buckets:     make(map[string]*Bucket),
		capacity:    capacity,
		refillEvery: refillEvery,
		now:         now,
	}
}

func (l *Limiter) Allow(key string) bool {
	return l.bucket(key).Allow()
}

func (l *Limiter) RetryAfter(key string) time.Duration {
	return l.bucket(key).RetryAfter()
}

func (l *Limiter) bucket(key string) *Bucket {
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, ok := l.buckets[key]
	if !ok {
		bucket = NewBucket(l.capacity, l.refillEvery, l.now)
		l.buckets[key] = bucket
	}
	return bucket
}
