package router

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Raphel6969/api-gateway/internal/config"
	"github.com/Raphel6969/api-gateway/internal/health"
	"github.com/Raphel6969/api-gateway/internal/loadbalancer"
	"github.com/Raphel6969/api-gateway/internal/proxy"
)

type Router struct {
	mux *http.ServeMux
}

func New(cfg *config.Config, log *slog.Logger) (*Router, error) {
	mux := http.NewServeMux()

	for _, r := range cfg.Routes {
		lb := loadbalancer.NewRoundRobin(r.Targets)

		checker := health.NewHealthCheck(r.Targets, lb, 3*time.Second, log)
		checker.Start()

		proxyHandler, err := proxy.NewLoadBalancedProxy(lb, r.Path)
		if err != nil {
			return nil, err
		}

		pathPattern := r.Path
		if !strings.HasSuffix(pathPattern, "/") {
			pathPattern = pathPattern + "/"
		}

		mux.Handle(pathPattern, proxyHandler)
	}

	return &Router{mux: mux}, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
