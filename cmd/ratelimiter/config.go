package main

import (
	"errors"
	"flag"
	"time"
)

type config struct {
	address     string
	capacity    int
	refillEvery time.Duration
	showVersion bool
}

func parseFlags(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("ratelimiter", flag.ContinueOnError)
	flags.StringVar(&cfg.address, "addr", ":8080", "address to listen on")
	flags.IntVar(&cfg.capacity, "capacity", 5, "maximum burst of requests allowed per client")
	flags.DurationVar(&cfg.refillEvery, "refill", time.Second, "interval at which one request is granted back to each client")
	flags.BoolVar(&cfg.showVersion, "version", false, "print the version and exit")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}

	if cfg.capacity < 1 {
		return config{}, errors.New("capacity must be at least 1")
	}
	if cfg.refillEvery <= 0 {
		return config{}, errors.New("refill interval must be positive")
	}
	return cfg, nil
}
