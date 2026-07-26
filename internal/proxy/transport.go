package proxy

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Raphel6969/api-gateway/internal/breaker"
)

type ResilientTransport struct {
	BaseTransport http.RoundTripper
	Breaker       *breaker.CircuitBreaker
	MaxRetries    int
	Log           *slog.Logger
}

func (t *ResilientTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.Breaker.Allow() {
		t.Log.Warn("Cirucit breaker OPEN , dropping Request", "path", req.URL.Path)
		return nil, breaker.ErrCircuitOption
	}

	var resp *http.Response
	var err error
	backoff := 100 * time.Millisecond

	// Making sure only safe to retry are going through
	isIdempotent := req.Method == http.MethodGet || req.Method == http.MethodHead || req.Method == http.MethodOptions

	// Set max attempts: retry idempotent requests, run mutating requests only once
	maxAttempts := t.MaxRetries
	if !isIdempotent {
		maxAttempts = 0
	}

	for attempt := 0; attempt <= maxAttempts; attempt++ {
		resp, err = t.BaseTransport.RoundTrip(req)

		if err == nil && resp.StatusCode < 500 {
			t.Breaker.RecordResult(true)
			return resp, nil
		}

		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		t.Log.Warn("Upstream request failed, retrying. . . ", "attempt", attempt+1, "path", req.URL.Path, "error", err)

		if attempt < t.MaxRetries {
			time.Sleep(backoff)
			backoff *= 2
		}
	}

	t.Breaker.RecordResult(false)
	return resp, err
}
