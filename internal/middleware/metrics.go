package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/Raphel6969/api-gateway/internal/metrics"
)

func Metrics(m *metrics.Metrics) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			m.InFlightRequests.Inc()
			defer m.InFlightRequests.Dec()

			start := time.Now()
			wrapped := NewResponseWriterInterceptor(w)

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start).Seconds()

			m.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
			m.RequestsTotal.WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(wrapped.statusCode)).Inc()
		})
	}
}
