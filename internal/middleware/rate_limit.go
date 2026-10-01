package middleware

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

func RateLimit(limiter Limiter, log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/metrics" || strings.HasPrefix(r.URL.Path, "/metrics") {
				next.ServeHTTP(w, r)
				return
			}

			clientKey := r.Header.Get(HeaderXUserID)
			if clientKey == "" {
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					clientKey = r.RemoteAddr
				} else {
					clientKey = host
				}
				clientKey = strings.Trim(clientKey, "[]")
			}

			allowed, err := limiter.Allow(r.Context(), clientKey)
			if err != nil {
				// Fail-open: upstream remains reachable if Redis has a transient hiccup
				log.Error("Rate limiter check failed, failing open", "error", err, "client_key", clientKey)
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				log.Warn("Rate limit exceeded", "client_key", clientKey, "path", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error": "Too Many Requests"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
