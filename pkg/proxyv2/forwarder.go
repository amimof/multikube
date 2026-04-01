package proxy

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
)

type Forwarder struct {
	transport http.RoundTripper
}

func NewForwarder(transport http.RoundTripper) *Forwarder {
	return &Forwarder{transport: transport}
}

func (f *Forwarder) Handler(pool *BackendPool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, ok := pool.Next(r)
		if !ok {
			http.Error(w, "no healthy upstream", http.StatusBadGateway)
			return
		}

		outReq := cloneRequestForTarget(r, target)
		if target.AuthInjector != nil {
			if err := target.AuthInjector.Apply(outReq); err != nil {
				writeProxyError(w, err)
				return
			}
		}
		resp, err := f.transport.RoundTrip(outReq)
		if err != nil {
			writeProxyError(w, err)
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		copyHeader(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)

		_, _ = io.Copy(flushWriter{ResponseWriter: w}, resp.Body)
	})
}

func cloneRequestForTarget(in *http.Request, target *BackendRuntime) *http.Request {
	out := in.Clone(in.Context())
	out.URL.Scheme = target.URL.Scheme
	out.URL.Host = target.URL.Host

	basePath := strings.TrimRight(target.URL.Path, "/")
	reqPath := in.URL.Path
	if basePath != "" {
		out.URL.Path = basePath + reqPath
	}

	out.Host = target.URL.Host
	out.RequestURI = ""

	copyForwardingHeaders(out, in)
	return out
}

func copyForwardingHeaders(out, in *http.Request) {
	out.Header = in.Header.Clone()

	remoteIP := clientIPFromRequest(in)
	appendHeader(out.Header, "X-Forwarded-For", remoteIP)
	out.Header.Set("X-Forwarded-Host", in.Host)

	if in.TLS != nil {
		out.Header.Set("X-Forwarded-Proto", "https")
	} else {
		out.Header.Set("X-Forwarded-Proto", "http")
	}
}

func appendHeader(h http.Header, key, value string) {
	if existing := h.Get(key); existing != "" {
		h.Set(key, existing+", "+value)
		return
	}
	h.Set(key, value)
}

func clientIPFromRequest(r *http.Request) string {
	hostPort := r.RemoteAddr
	host, _, err := net.SplitHostPort(hostPort)
	if err == nil {
		return host
	}
	return hostPort
}

func copyHeader(dst, src http.Header) {
	for k, values := range src {
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}

func writeProxyError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
		return
	}
	http.Error(w, "upstream request failed", http.StatusBadGateway)
}

type flushWriter struct {
	http.ResponseWriter
}

func (fw flushWriter) Write(p []byte) (int, error) {
	n, err := fw.ResponseWriter.Write(p)
	if f, ok := fw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
	return n, err
}
