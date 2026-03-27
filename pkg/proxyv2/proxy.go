package proxy

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

type Proxy struct {
	runtime *RuntimeStore
}

func NewProxy(runtime *RuntimeStore) *Proxy {
	return &Proxy{runtime: runtime}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt := p.runtime.Load()

	route, ok := rt.Match(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	handler := route.Handler
	if route.Timeout > 0 {
		handler = timeoutMiddleware(route.Timeout)(handler)
	}

	handler = withRuntimeVersion(rt.Version)(handler)
	handler.ServeHTTP(w, r)
}

func timeoutMiddleware(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func withRuntimeVersion(version uint64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Proxy-Config-Version", itoa(version))
			next.ServeHTTP(w, r)
		})
	}
}

func itoa(v uint64) string {
	return strconv.FormatUint(v, 10)
}
