package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rr-saberov/ratelimiter/internal/httpapi"
	"github.com/rr-saberov/ratelimiter/ratelimit"
)

func newHandler(capacity int, refillEvery time.Duration) http.Handler {
	fixedNow := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	limiter := ratelimit.NewLimiter(capacity, refillEvery, func() time.Time { return fixedNow })
	return httpapi.NewHandler(limiter)
}

func get(t *testing.T, handler http.Handler, remoteAddress string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", http.NoBody)
	request.RemoteAddr = remoteAddress
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestHandlerServesRequestsWithinLimit(t *testing.T) {
	handler := newHandler(2, time.Second)

	for i := range 2 {
		response := get(t, handler, "10.0.0.1:40001")
		if response.Code != http.StatusOK {
			t.Fatalf("request %d status: want %d, got %d", i+1, http.StatusOK, response.Code)
		}
		if body := response.Body.String(); body != "ok\n" {
			t.Fatalf("request %d body: want %q, got %q", i+1, "ok\n", body)
		}
	}
}

func TestHandlerRejectsRequestOverLimitWithRetryAfter(t *testing.T) {
	handler := newHandler(1, time.Second)
	get(t, handler, "10.0.0.1:40001")

	response := get(t, handler, "10.0.0.1:40001")

	if response.Code != http.StatusTooManyRequests {
		t.Fatalf("status: want %d, got %d", http.StatusTooManyRequests, response.Code)
	}
	if got := response.Header().Get("Retry-After"); got != "1" {
		t.Fatalf("Retry-After: want %q, got %q", "1", got)
	}
}

func TestHandlerRoundsRetryAfterUpToWholeSeconds(t *testing.T) {
	handler := newHandler(1, 1500*time.Millisecond)
	get(t, handler, "10.0.0.1:40001")

	response := get(t, handler, "10.0.0.1:40001")

	if got := response.Header().Get("Retry-After"); got != "2" {
		t.Fatalf("Retry-After: want %q, got %q", "2", got)
	}
}

func TestHandlerLimitsEachClientAddressSeparately(t *testing.T) {
	handler := newHandler(1, time.Second)
	get(t, handler, "10.0.0.1:40001")

	if response := get(t, handler, "10.0.0.2:40002"); response.Code != http.StatusOK {
		t.Fatalf("other client: want %d, got %d", http.StatusOK, response.Code)
	}
	if response := get(t, handler, "10.0.0.1:40003"); response.Code != http.StatusTooManyRequests {
		t.Fatalf("same client from a new port: want %d, got %d", http.StatusTooManyRequests, response.Code)
	}
}
