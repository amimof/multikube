package proxy

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"time"
)

// responseWriter implements http.ResponseWriter and adds status code and response length bytes
// so that WithLogging middleware can log response status codes
type responseWriter struct {
	http.ResponseWriter
	status int
	length int
}

// WriteHeader sends and sets an HTTP response header with the provided
// status code. Implements the http.ResponseWriter interface
func (r *responseWriter) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write implements the http.ResponseWriter interface
func (r *responseWriter) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = 200
	}
	n, err := r.ResponseWriter.Write(b)
	r.length += n
	return n, err
}

// Hijack implements the http.Hijacker interface
func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if r.length < 0 {
		r.length = 0
	}
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

// WithLogging applies access log style logging to the HTTP server
func WithLogging() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			lrw := &responseWriter{ResponseWriter: w}
			next.ServeHTTP(lrw, r)
			var isResCached bool
			if lrw.Header().Get("Multikube-Cache-Age") != "" {
				isResCached = true
			}
			duration := time.Now().Sub(start)
			log.Printf("%s %s %s %s %s %d %d %s %t", r.Method, r.URL.Path, r.URL.RawQuery, r.RemoteAddr, r.Proto, lrw.status, lrw.length, duration.String(), isResCached)
		})
	}
}
