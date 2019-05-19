package gateway

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Proxy pipes the traffic between the requester (Prometheus / Grafana) and upstream service
type Proxy struct {
	targetAddress *url.URL
	targetName    string

	reverseProxy *httputil.ReverseProxy
}

// newProxy creates a new reverse proxy for a single upstream service
func newProxy(target string, targetName string) (*Proxy, error) {
	url, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		targetAddress: url,
		targetName:    targetName,
		reverseProxy: &httputil.ReverseProxy{
			Director: newDirector(url),
		},
	}, nil
}

// Handler is the route handler which must be bound to the routes
func (p *Proxy) Handler(res http.ResponseWriter, req *http.Request) {
	// This launches a new Go routine under the hood and therefore it's non blocking
	p.reverseProxy.ServeHTTP(res, req)
}

func newDirector(targetURL *url.URL) func(req *http.Request) {
	targetQuery := targetURL.RawQuery

	return func(req *http.Request) {
		// Update headers to support SSL redirection
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
		req.URL.Path = singleJoiningSlash(targetURL.Path, req.URL.Path)
		req.Host = targetURL.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
}
