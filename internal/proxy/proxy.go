package proxy

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/Raphel6969/api-gateway/internal/breaker"
	"github.com/Raphel6969/api-gateway/internal/loadbalancer"
)

// NewLoadBalancedProxy creates a proxy that picks a dynamic backend for every request.
func NewLoadBalancedProxy(lb loadbalancer.LoadBalancer, pathPrefix string, log *slog.Logger) (http.Handler, error) {
	cb := breaker.NewCircuitBreaker(3, 10*time.Second)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Pick next target from Load Balancer
		targetURL, err := lb.Next()
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`{"error": "Service Unavailable"}`))
			return
		}

		target, err := url.Parse(targetURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 2. Build reverse proxy for selected target
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.Transport = &ResilientTransport{
			BaseTransport: http.DefaultTransport,
			Breaker:       cb,
			MaxRetries:    2,
			Log:           log,
		}

		originalDirector := proxy.Director

		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			req.Host = target.Host

			if pathPrefix != "" {
				req.URL.Path = strings.TrimPrefix(req.URL.Path, pathPrefix)
				if !strings.HasPrefix(req.URL.Path, "/") {
					req.URL.Path = "/" + req.URL.Path
				}
			}

			if clientIP := req.Header.Get("X-Forwarded-For"); clientIP == "" {
				req.Header.Set("X-Forwarded-For", req.RemoteAddr)
			}
			req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		}

		// 3. Serve request
		proxy.ServeHTTP(w, r)
	}), nil
}
