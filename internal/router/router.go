package router

import (
	"net/http"
	"strings"

	"github.com/Raphel6969/api-gateway/internal/config"
	"github.com/Raphel6969/api-gateway/internal/proxy"
)

type Router struct {
	mux *http.ServeMux
}

func New(cfg *config.Config) (*Router, error) {
	mux := http.NewServeMux()

	for _, r := range cfg.Routes {
		proxyHandler, err := proxy.NewReverseProxy(r.Target, r.Path)
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
