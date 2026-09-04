package httpapi

import (
	"io"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Limiter interface {
	Allow(key string) bool
	RetryAfter(key string) time.Duration
}

func NewHandler(limiter Limiter) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", serveOK)
	return Limit(limiter, mux)
}

func Limit(limiter Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := clientKey(r)
		if !limiter.Allow(key) {
			w.Header().Set("Retry-After", retryAfterSeconds(limiter.RetryAfter(key)))
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func serveOK(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, "ok\n")
}

func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func retryAfterSeconds(d time.Duration) string {
	return strconv.Itoa(int(math.Ceil(d.Seconds())))
}
