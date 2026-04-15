package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
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

// ---------------------------------------------------------------------------
// Tests — stripImpersonationHeaders
// ---------------------------------------------------------------------------

func TestStripImpersonationHeaders(t *testing.T) {
	req := newRequest(t, "GET", "https://proxy:8443/api/v1/pods")
	req.Header.Set("Impersonate-User", "evil-user")
	req.Header.Set("Impersonate-Group", "admin")
	req.Header.Set("Impersonate-Extra-Scopes", "cluster-admin")
	req.Header.Set("Authorization", "Bearer keep-me")

	stripImpersonationHeaders(req)

	if req.Header.Get("Impersonate-User") != "" {
		t.Error("Impersonate-User should be stripped")
	}
	if req.Header.Get("Impersonate-Group") != "" {
		t.Error("Impersonate-Group should be stripped")
	}
	if req.Header.Get("Impersonate-Extra-Scopes") != "" {
		t.Error("Impersonate-Extra-Scopes should be stripped")
	}
	if req.Header.Get("Authorization") != "Bearer keep-me" {
		t.Error("Authorization header should be preserved")
	}
}

// ---------------------------------------------------------------------------
// Tests — injectImpersonationHeaders
// ---------------------------------------------------------------------------

func TestInjectImpersonation_DefaultClaims(t *testing.T) {
	principal := &Principal{
		Subject: "alice",
		Groups:  []string{"devs", "admins"},
		Claims:  map[string]any{"sub": "alice", "groups": []any{"devs", "admins"}},
	}
	cfg := &ImpersonationRuntime{
		Enabled:       true,
		UsernameClaim: "sub",
		GroupsClaim:   "groups",
	}

	req := newRequest(t, "GET", "https://upstream:6443/api/v1/pods")
	err := injectImpersonationHeaders(req, principal, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := req.Header.Get("Impersonate-User"); got != "alice" {
		t.Errorf("Impersonate-User = %q, want %q", got, "alice")
	}
	groups := req.Header.Values("Impersonate-Group")
	if len(groups) != 2 || groups[0] != "devs" || groups[1] != "admins" {
		t.Errorf("Impersonate-Group = %v, want [devs admins]", groups)
	}
}

func TestInjectImpersonation_CustomUsernameClaim(t *testing.T) {
	principal := &Principal{
		Subject: "sub-id-123",
		Claims:  map[string]any{"email": "alice@example.com", "sub": "sub-id-123"},
	}
	cfg := &ImpersonationRuntime{
		Enabled:       true,
		UsernameClaim: "email",
		GroupsClaim:   "groups",
	}

	req := newRequest(t, "GET", "https://upstream:6443/api/v1/pods")
	err := injectImpersonationHeaders(req, principal, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := req.Header.Get("Impersonate-User"); got != "alice@example.com" {
		t.Errorf("Impersonate-User = %q, want %q", got, "alice@example.com")
	}
}

func TestInjectImpersonation_CustomGroupsClaim(t *testing.T) {
	principal := &Principal{
		Subject: "alice",
		Groups:  []string{"default-group"},
		Claims: map[string]any{
			"sub":   "alice",
			"roles": []any{"editor", "viewer"},
		},
	}
	cfg := &ImpersonationRuntime{
		Enabled:       true,
		UsernameClaim: "sub",
		GroupsClaim:   "roles",
	}

	req := newRequest(t, "GET", "https://upstream:6443/api/v1/pods")
	err := injectImpersonationHeaders(req, principal, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	groups := req.Header.Values("Impersonate-Group")
	if len(groups) != 2 || groups[0] != "editor" || groups[1] != "viewer" {
		t.Errorf("Impersonate-Group = %v, want [editor viewer]", groups)
	}
}

func TestInjectImpersonation_ExtraClaims(t *testing.T) {
	principal := &Principal{
		Subject: "alice",
		Claims: map[string]any{
			"sub":    "alice",
			"scopes": []any{"read", "write"},
			"tenant": "acme",
		},
	}
	cfg := &ImpersonationRuntime{
		Enabled:       true,
		UsernameClaim: "sub",
		GroupsClaim:   "groups",
		ExtraClaims:   []string{"scopes", "tenant", "missing-claim"},
	}

	req := newRequest(t, "GET", "https://upstream:6443/api/v1/pods")
	err := injectImpersonationHeaders(req, principal, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	scopes := req.Header.Values("Impersonate-Extra-scopes")
	if len(scopes) != 2 || scopes[0] != "read" || scopes[1] != "write" {
		t.Errorf("Impersonate-Extra-scopes = %v, want [read write]", scopes)
	}

	tenant := req.Header.Values("Impersonate-Extra-tenant")
	if len(tenant) != 1 || tenant[0] != "acme" {
		t.Errorf("Impersonate-Extra-tenant = %v, want [acme]", tenant)
	}

	// Missing claim should not produce a header.
	if req.Header.Get("Impersonate-Extra-missing-claim") != "" {
		t.Error("Impersonate-Extra-missing-claim should not be set for missing claim")
	}
}

func TestInjectImpersonation_MissingUsernameClaim_Error(t *testing.T) {
	principal := &Principal{
		Subject: "",
		Claims:  map[string]any{},
	}
	cfg := &ImpersonationRuntime{
		Enabled:       true,
		UsernameClaim: "sub",
		GroupsClaim:   "groups",
	}

	req := newRequest(t, "GET", "https://upstream:6443/api/v1/pods")
	err := injectImpersonationHeaders(req, principal, cfg)
	if err == nil {
		t.Fatal("expected error for empty subject claim")
	}
}

func TestInjectImpersonation_CustomUsernameClaim_Missing_Error(t *testing.T) {
	principal := &Principal{
		Subject: "alice",
		Claims:  map[string]any{"sub": "alice"},
	}
	cfg := &ImpersonationRuntime{
		Enabled:       true,
		UsernameClaim: "email", // not in claims
		GroupsClaim:   "groups",
	}

	req := newRequest(t, "GET", "https://upstream:6443/api/v1/pods")
	err := injectImpersonationHeaders(req, principal, cfg)
	if err == nil {
		t.Fatal("expected error for missing custom username claim")
	}
}

func TestInjectImpersonation_GroupsClaimString(t *testing.T) {
	// When groups claim is a single string rather than an array.
	principal := &Principal{
		Subject: "alice",
		Claims: map[string]any{
			"sub":  "alice",
			"role": "admin",
		},
	}
	cfg := &ImpersonationRuntime{
		Enabled:       true,
		UsernameClaim: "sub",
		GroupsClaim:   "role",
	}

	req := newRequest(t, "GET", "https://upstream:6443/api/v1/pods")
	err := injectImpersonationHeaders(req, principal, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	groups := req.Header.Values("Impersonate-Group")
	if len(groups) != 1 || groups[0] != "admin" {
		t.Errorf("Impersonate-Group = %v, want [admin]", groups)
	}
}

// ---------------------------------------------------------------------------
// Tests — Forwarder.Handler impersonation integration
// ---------------------------------------------------------------------------

func TestForwarderHandler_ImpersonationEnabled_NoPrincipal_Returns403(t *testing.T) {
	// Simulate a backend pool with impersonation enabled but no principal on context.
	pool := &BackendPool{
		Name: "test",
		Targets: []*BackendRuntime{
			{
				URL: mustParseURL(t, "https://upstream:6443"),
			},
		},
		Iterator: &RoundRobinLB{},
		Impersonation: &ImpersonationRuntime{
			Enabled:       true,
			UsernameClaim: "sub",
			GroupsClaim:   "groups",
		},
	}

	fwd := NewForwarder(http.DefaultTransport)
	handler := fwd.Handler(pool)

	req := newRequest(t, "GET", "https://proxy:8443/api/v1/pods")
	// No principal on context — should be rejected.
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
}

func TestForwarderHandler_ImpersonationDisabled_NoPrincipal_Allowed(t *testing.T) {
	// When impersonation is disabled, requests without a principal should pass
	// through to the upstream (we use a test server to verify).
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify no impersonation headers are set
		if r.Header.Get("Impersonate-User") != "" {
			t.Error("Impersonate-User should not be set when impersonation disabled")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	pool := &BackendPool{
		Name: "test",
		Targets: []*BackendRuntime{
			{
				URL: mustParseURL(t, upstream.URL),
			},
		},
		Iterator: &RoundRobinLB{},
		Impersonation: &ImpersonationRuntime{
			Enabled: false,
		},
	}

	fwd := NewForwarder(http.DefaultTransport)
	handler := fwd.Handler(pool)

	req := newRequest(t, "GET", "https://proxy:8443/api/v1/pods")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestForwarderHandler_ClientImpersonationHeadersStripped(t *testing.T) {
	// Even when impersonation is disabled, client-supplied Impersonate-*
	// headers must be stripped.
	var receivedHeaders http.Header
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	pool := &BackendPool{
		Name: "test",
		Targets: []*BackendRuntime{
			{
				URL: mustParseURL(t, upstream.URL),
			},
		},
		Iterator: &RoundRobinLB{},
		Impersonation: &ImpersonationRuntime{
			Enabled: false,
		},
	}

	fwd := NewForwarder(http.DefaultTransport)
	handler := fwd.Handler(pool)

	req := newRequest(t, "GET", "https://proxy:8443/api/v1/pods")
	req.Header.Set("Impersonate-User", "evil-user")
	req.Header.Set("Impersonate-Group", "admin")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if receivedHeaders.Get("Impersonate-User") != "" {
		t.Error("client-supplied Impersonate-User should have been stripped")
	}
	if receivedHeaders.Get("Impersonate-Group") != "" {
		t.Error("client-supplied Impersonate-Group should have been stripped")
	}
}
