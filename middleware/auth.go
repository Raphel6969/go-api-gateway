package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/Raphel6969/api-gateway/internal/auth"
)

const (
	HeaderXUserID   = "X-User-ID"
	HeaderXUserRole = "X-User-Role"
)

func Authenticate(authenticator auth.Authenticator, publicPaths []string, log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range publicPaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}

			if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
				userID, valid := authenticator.ValidateAPIKey(apiKey)
				if valid {
					r.Header.Set(HeaderXUserID, userID)
					r.Header.Set(HeaderXUserRole, "api_client")
					next.ServeHTTP(w, r)
					return
				}
			}

			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := authenticator.ValidateJWT(tokenStr)
				if err == nil {
					r.Header.Set(HeaderXUserID, claims.UserID)
					r.Header.Set(HeaderXUserRole, claims.Role)
					next.ServeHTTP(w, r)
					return
				}
			}

			log.Warn("Unauthorized request blocked", "path", r.URL.Path, "remote_addr", r.RemoteAddr)
			w.Header().Set("Contet-Type", "applicaton/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "Unauthorized"}`))
		})
	}
}
