package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"time"

	"github.com/Raphel6969/api-gateway/internal/cache"
)

type cacheResponseWrite struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (c *cacheResponseWrite) WriteHeader(code int) {
	c.statusCode = code
	c.ResponseWriter.WriteHeader(code)
}

func (c *cacheResponseWrite) Write(b []byte) (int, error) {
	c.body.Write(b)
	return c.ResponseWriter.Write(b)
}

func CacheMiddleware(c *cache.MemoryCache, ttl time.Duration, log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			userID := r.Header.Get("X-User-ID")
			cacheKey := r.URL.Path + ":" + userID

			if cachedBody, statusCode, found := c.Get(cacheKey); found {
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(statusCode)
				w.Write(cachedBody)
				return
			}

			crw := &cacheResponseWrite{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}
			w.Header().Set("X-Cache", "MISS")
			next.ServeHTTP(w, r)

			if crw.statusCode == http.StatusOK {
				c.Set(cacheKey, crw.statusCode, crw.body.Bytes(), ttl)
			}
		})
	}
}
