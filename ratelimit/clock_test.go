package ratelimit_test

import "time"

type fakeClock struct {
	current time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{current: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) Now() time.Time {
	return c.current
}

func (c *fakeClock) Advance(d time.Duration) {
	c.current = c.current.Add(d)
}
