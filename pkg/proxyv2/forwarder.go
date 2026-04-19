package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Forwarder struct {
	transport http.RoundTripper
	metrics   *ProxyMetrics
}

func NewForwarder(transport http.RoundTripper) *Forwarder {
	return &Forwarder{transport: transport}
}

func NewForwarderWithMetrics(transport http.RoundTripper, metrics *ProxyMetrics) *Forwarder {
	return &Forwarder{transport: transport, metrics: metrics}
}

func (f *Forwarder) Handler(pool *BackendPool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, ok := pool.Next(r)
		if !ok {
			http.Error(w, "no healthy upstream", http.StatusBadGateway)
			return
		}

		backendName := target.Name
		ctx := r.Context()

		// Track backend active requests
		if f.metrics != nil {
			backendAttr := metric.WithAttributes(attribute.String("backend", backendName))
			f.metrics.BackendActiveRequests.Add(ctx, 1, backendAttr)
			defer f.metrics.BackendActiveRequests.Add(ctx, -1, backendAttr)
		}

		outReq := cloneRequestForTarget(r, target)

		// Always strip client-supplied impersonation headers — the proxy
		// owns these exclusively when impersonation is configured.
		stripImpersonationHeaders(outReq)

		// Impersonation header injection.
		if pool.Impersonation != nil && pool.Impersonation.Enabled {
			principal, hasPrincipal := PrincipalFromContext(r.Context())
			if !hasPrincipal || principal == nil {
				http.Error(w, "Forbidden: authentication required for impersonation", http.StatusForbidden)
				return
			}
			if err := injectImpersonationHeaders(outReq, principal, pool.Impersonation); err != nil {
				http.Error(w, "Forbidden: "+err.Error(), http.StatusForbidden)
				return
			}
		}

		if target.AuthInjector != nil {
			if err := target.AuthInjector.Apply(outReq); err != nil {
				writeProxyError(w, err)
				return
			}
		}

		start := time.Now()
		resp, err := f.transport.RoundTrip(outReq)
		if err != nil {
			if f.metrics != nil {
				f.metrics.BackendRequestsTotal.Add(ctx, 1,
					metric.WithAttributes(
						attribute.String("backend", backendName),
						attribute.Int("status_code", 502),
					))
				f.metrics.BackendRequestDuration.Record(ctx, time.Since(start).Seconds(),
					metric.WithAttributes(attribute.String("backend", backendName)))
			}
			writeProxyError(w, err)
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		if f.metrics != nil {
			f.metrics.BackendRequestsTotal.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("backend", backendName),
					attribute.Int("status_code", resp.StatusCode),
				))
			f.metrics.BackendRequestDuration.Record(ctx, time.Since(start).Seconds(),
				metric.WithAttributes(attribute.String("backend", backendName)))
		}

		copyHeader(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)

		_, _ = io.Copy(flushWriter{ResponseWriter: w}, resp.Body)
	})
}

func cloneRequestForTarget(in *http.Request, target *BackendRuntime) *http.Request {
	out := in.Clone(in.Context())
	out.URL.Scheme = target.URL.Scheme
	out.URL.Host = target.URL.Host

	reqPath := stripMatchedPath(in)

	basePath := strings.TrimRight(target.URL.Path, "/")
	if basePath != "" {
		out.URL.Path = basePath + reqPath
	} else {
		out.URL.Path = reqPath
	}

	out.Host = target.URL.Host
	out.RequestURI = ""

	copyForwardingHeaders(out, in)
	return out
}

// stripMatchedPath removes the matched route prefix (or exact path) from the
// incoming request path so the upstream sees a path relative to its own root.
// For non-path route kinds (header, JWT, SNI) the original path is returned
// unchanged.
func stripMatchedPath(in *http.Request) string {
	route, ok := MatchedRouteFromContext(in.Context())
	if !ok {
		return in.URL.Path
	}

	reqPath := in.URL.Path

	switch route.Kind {
	case RouteMatchKindPathPrefix:
		reqPath = strings.TrimPrefix(reqPath, route.PathPrefix)
	case RouteMatchKindPath:
		reqPath = strings.TrimPrefix(reqPath, route.Path)
	default:
		return reqPath
	}

	// Guarantee the upstream path is absolute.
	if reqPath == "" || reqPath[0] != '/' {
		reqPath = "/" + reqPath
	}

	return reqPath
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

// stripImpersonationHeaders removes all client-supplied Impersonate-* headers
// from the outbound request. The proxy owns these headers exclusively.
func stripImpersonationHeaders(req *http.Request) {
	for key := range req.Header {
		if strings.HasPrefix(strings.ToLower(key), "impersonate-") {
			req.Header.Del(key)
		}
	}
}

// injectImpersonationHeaders sets Kubernetes impersonation headers on the
// outbound request using the authenticated principal and the backend's
// impersonation configuration.
func injectImpersonationHeaders(req *http.Request, principal *Principal, cfg *ImpersonationRuntime) error {
	// --- Impersonate-User ---
	username, err := resolveUsername(principal, cfg.UsernameClaim)
	if err != nil {
		return err
	}
	req.Header.Set("Impersonate-User", username)

	// --- Impersonate-Group ---
	groups := resolveGroups(principal, cfg.GroupsClaim)
	for _, g := range groups {
		req.Header.Add("Impersonate-Group", g)
	}

	// --- Impersonate-Extra-<claim> ---
	for _, claim := range cfg.ExtraClaims {
		vals := resolveClaimValues(principal, claim)
		if len(vals) == 0 {
			continue // skip missing extra claims
		}
		headerName := "Impersonate-Extra-" + claim
		for _, v := range vals {
			req.Header.Add(headerName, v)
		}
	}

	return nil
}

// resolveUsername extracts the username string from the principal based on the
// configured claim name.
func resolveUsername(principal *Principal, claim string) (string, error) {
	if claim == "sub" {
		if principal.Subject == "" {
			return "", fmt.Errorf("username claim %q is empty", claim)
		}
		return principal.Subject, nil
	}
	v, ok := principal.Claims[claim]
	if !ok {
		return "", fmt.Errorf("username claim %q not found in JWT", claim)
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", fmt.Errorf("username claim %q is not a non-empty string", claim)
	}
	return s, nil
}

// resolveGroups extracts group strings from the principal based on the
// configured claim name. Returns nil if the claim is missing.
func resolveGroups(principal *Principal, claim string) []string {
	if claim == "groups" {
		return principal.Groups
	}
	v, ok := principal.Claims[claim]
	if !ok {
		return nil
	}
	return toStringSlice(v)
}

// resolveClaimValues extracts string values for a single claim, supporting
// both scalar string and array-of-string JWT claim shapes.
func resolveClaimValues(principal *Principal, claim string) []string {
	v, ok := principal.Claims[claim]
	if !ok {
		return nil
	}
	return toStringSlice(v)
}
