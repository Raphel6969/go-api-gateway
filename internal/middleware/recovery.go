package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recovery(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := string(debug.Stack())

					log.Error(
						"Unhandled panic recovered",
						"error", err,
						"stack", stack,
						"path", r.URL.Path,
					)

					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error": Internal Server Error}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
