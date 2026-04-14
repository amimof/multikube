package proxy

import (
	"context"
	"net/http"
	"net/url"
	"testing"
)

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("bad url %q: %v", raw, err)
	}
	return u
}

func newRequest(t *testing.T, method, rawURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, rawURL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	return req
}

func TestStripMatchedPath_PathPrefix(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		reqPath    string
		targetBase string // backend URL path (e.g. "/k8s")
		wantPath   string
	}{
		{
			name:     "exact prefix stripped to root",
			prefix:   "/demo",
			reqPath:  "/demo",
			wantPath: "/",
		},
		{
			name:     "prefix with trailing slash stripped to root",
			prefix:   "/demo",
			reqPath:  "/demo/",
			wantPath: "/",
		},
		{
			name:     "nested path after prefix",
			prefix:   "/demo",
			reqPath:  "/demo/api/v1/pods",
			wantPath: "/api/v1/pods",
		},
		{
			name:     "multi-segment prefix",
			prefix:   "/clusters/dev",
			reqPath:  "/clusters/dev/api/v1/nodes",
			wantPath: "/api/v1/nodes",
		},
		{
			name:       "backend base path combined with stripped path",
			prefix:     "/demo",
			reqPath:    "/demo/foo",
			targetBase: "/k8s",
			wantPath:   "/k8s/foo",
		},
		{
			name:       "backend base path combined with root after strip",
			prefix:     "/demo",
			reqPath:    "/demo",
			targetBase: "/k8s",
			wantPath:   "/k8s/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &RouteRuntime{
				Kind:       RouteMatchKindPathPrefix,
				PathPrefix: tt.prefix,
			}
			ctx := WithMatchedRoute(context.Background(), route)
			req := newRequest(t, "GET", "https://proxy:8443"+tt.reqPath)
			req = req.WithContext(ctx)

			targetURL := "https://upstream:6443"
			if tt.targetBase != "" {
				targetURL += tt.targetBase
			}
			target := &BackendRuntime{
				URL: mustParseURL(t, targetURL),
			}

			out := cloneRequestForTarget(req, target)

			if out.URL.Path != tt.wantPath {
				t.Errorf("path = %q, want %q", out.URL.Path, tt.wantPath)
			}
		})
	}
}

func TestStripMatchedPath_ExactPath(t *testing.T) {
	tests := []struct {
		name       string
		matchPath  string
		reqPath    string
		targetBase string
		wantPath   string
	}{
		{
			name:      "exact path stripped to root",
			matchPath: "/api/v1/nodes",
			reqPath:   "/api/v1/nodes",
			wantPath:  "/",
		},
		{
			name:       "exact path with backend base",
			matchPath:  "/api/v1/nodes",
			reqPath:    "/api/v1/nodes",
			targetBase: "/k8s",
			wantPath:   "/k8s/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &RouteRuntime{
				Kind: RouteMatchKindPath,
				Path: tt.matchPath,
			}
			ctx := WithMatchedRoute(context.Background(), route)
			req := newRequest(t, "GET", "https://proxy:8443"+tt.reqPath)
			req = req.WithContext(ctx)

			targetURL := "https://upstream:6443"
			if tt.targetBase != "" {
				targetURL += tt.targetBase
			}
			target := &BackendRuntime{
				URL: mustParseURL(t, targetURL),
			}

			out := cloneRequestForTarget(req, target)

			if out.URL.Path != tt.wantPath {
				t.Errorf("path = %q, want %q", out.URL.Path, tt.wantPath)
			}
		})
	}
}

func TestStripMatchedPath_NonPathRouteUnchanged(t *testing.T) {
	tests := []struct {
		name string
		kind RouteMatchKind
	}{
		{"header", RouteMatchKindHeader},
		{"SNI", RouteMatchKindSNI},
		{"JWT", RouteMatchKindJWT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := &RouteRuntime{Kind: tt.kind}
			ctx := WithMatchedRoute(context.Background(), route)
			req := newRequest(t, "GET", "https://proxy:8443/api/v1/pods")
			req = req.WithContext(ctx)

			target := &BackendRuntime{
				URL: mustParseURL(t, "https://upstream:6443"),
			}

			out := cloneRequestForTarget(req, target)

			if out.URL.Path != "/api/v1/pods" {
				t.Errorf("path = %q, want %q", out.URL.Path, "/api/v1/pods")
			}
		})
	}
}

func TestStripMatchedPath_NoRouteOnContext(t *testing.T) {
	// When no matched route is on the context (should not happen in
	// practice), the original path must be forwarded unchanged.
	req := newRequest(t, "GET", "https://proxy:8443/anything")
	target := &BackendRuntime{
		URL: mustParseURL(t, "https://upstream:6443"),
	}

	out := cloneRequestForTarget(req, target)

	if out.URL.Path != "/anything" {
		t.Errorf("path = %q, want %q", out.URL.Path, "/anything")
	}
}

func TestStripMatchedPath_QueryStringPreserved(t *testing.T) {
	route := &RouteRuntime{
		Kind:       RouteMatchKindPathPrefix,
		PathPrefix: "/demo",
	}
	ctx := WithMatchedRoute(context.Background(), route)
	req := newRequest(t, "GET", "https://proxy:8443/demo/api/v1/pods?watch=true&limit=100")
	req = req.WithContext(ctx)

	target := &BackendRuntime{
		URL: mustParseURL(t, "https://upstream:6443"),
	}

	out := cloneRequestForTarget(req, target)

	if out.URL.Path != "/api/v1/pods" {
		t.Errorf("path = %q, want %q", out.URL.Path, "/api/v1/pods")
	}
	if out.URL.RawQuery != "watch=true&limit=100" {
		t.Errorf("query = %q, want %q", out.URL.RawQuery, "watch=true&limit=100")
	}
}
