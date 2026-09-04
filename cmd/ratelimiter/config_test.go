package main

import (
	"testing"
	"time"
)

func TestParseFlagsUsesDefaults(t *testing.T) {
	cfg, err := parseFlags(nil)
	if err != nil {
		t.Fatalf("parseFlags: unexpected error: %v", err)
	}

	want := config{address: ":8080", capacity: 5, refillEvery: time.Second}
	if cfg != want {
		t.Fatalf("config: want %+v, got %+v", want, cfg)
	}
}

func TestParseFlagsReadsOverrides(t *testing.T) {
	cfg, err := parseFlags([]string{"-addr", ":9090", "-capacity", "20", "-refill", "250ms", "-version"})
	if err != nil {
		t.Fatalf("parseFlags: unexpected error: %v", err)
	}

	want := config{address: ":9090", capacity: 20, refillEvery: 250 * time.Millisecond, showVersion: true}
	if cfg != want {
		t.Fatalf("config: want %+v, got %+v", want, cfg)
	}
}

func TestParseFlagsRejectsNonPositiveCapacity(t *testing.T) {
	if _, err := parseFlags([]string{"-capacity", "0"}); err == nil {
		t.Fatal("capacity 0: want error, got nil")
	}
}

func TestParseFlagsRejectsNonPositiveRefillInterval(t *testing.T) {
	if _, err := parseFlags([]string{"-refill", "0s"}); err == nil {
		t.Fatal("refill 0s: want error, got nil")
	}
}
