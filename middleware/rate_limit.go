package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/Raphel6969/api-gateway/internal/ratelimit"
)

func RateLimit(limiter *ratelimit.RateLimiter, log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientKey := r.Header.Get(HeaderXUserID)
			if clientKey == "" {
				clientKey = strings.Split(r.RemoteAddr, ":")[0]
			}

			if !limiter.Allow(clientKey) {
				log.Warn("Rate limit exceeded", "client_key", clientKey, "path", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error": "Too Many Request"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
