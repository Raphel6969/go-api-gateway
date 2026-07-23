package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewReverseProxy(targetURL string, pathPrefix string) (http.Handler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

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

		if clientIP := req.Header.Get("X-forwarded-For"); clientIP == "" {
			req.Header.Set("X-forwarded-For", req.RemoteAddr)
		}
		req.Header.Set("X-forwarded-For", req.Header.Get("Host"))
	}

	return proxy, nil
}
