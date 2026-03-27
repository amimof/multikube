package compile_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
	"github.com/amimof/multikube/pkg/compile"
)

// ---------------------------------------------------------------------------
// Test fixtures
// ---------------------------------------------------------------------------

// testCertPEM and testKeyPEM are a self-signed RSA cert/key pair generated
// exclusively for use in tests. They must NOT be used in production.
const testCertPEM = `-----BEGIN CERTIFICATE-----
MIIC/zCCAeegAwIBAgIUGAAq1t3SR5mAIjPTmVYx1J5nrAcwDQYJKoZIhvcNAQEL
BQAwDzENMAsGA1UEAwwEdGVzdDAeFw0yNjAzMjUxMzQ1MDhaFw0zNjAzMjIxMzQ1
MDhaMA8xDTALBgNVBAMMBHRlc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQDIlkHiaQ+lngtv5bXd9I1rCqPkLyB1ksjIwi+O0eX8nfxrEoTs1SfZLbUW
pEZpaKwN1u3aiDl2Yz2rYqBGiJvgRTU+xhS7Hx3ZEgk5WhnC0y6bpV1xgIQSiKkp
PANiTJupEkS/3V/RSeKgy1R1ECxl3lduT8mO2fu+xcCpaMBABtxUvwm/Day6keE5
8igoL8XzfRnGvlUIBzYbEuUJ4NnH4UobYCuWoqYhPCyjI7rMN4tcFrClS0oAfCKa
opdD/NpUdfbywiiO9lelK4BNN4UBZfNmVYQdBbo/PA8Itd5uxisaok/BI7DS7HNx
rhDCli3gZOtqFHyH8wKA50DZi4oBAgMBAAGjUzBRMB0GA1UdDgQWBBSlXG77Sks0
hmvPvTa8/TVjCs9rHTAfBgNVHSMEGDAWgBSlXG77Sks0hmvPvTa8/TVjCs9rHTAP
BgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQB9NccpK78d4Aq3Ab2c
EFGty1K94ZAiBmvJNZgTUppc0nPknKfXHtzQnTSxgKE/7Lzfz9iKnZCpLWod5QHf
OCdPNwTgxyfFNohDHeVkgY2lIdY5chcNy0lP+EInSB+adCy2aDWanb9/qygVjpyc
W/8EKScxPzpeGt4PYQF7qhmAet82cMivY1Dma/7FK4pjd0UKsooXIdfj5afnMsz6
U5oue90en93Y2LdLPdkCKbP/n5q+2JicbsqD8ctbCaOPXQ8XDsmZ2FlR7nnzOTdA
zyUdWOm3uUGN+hHRFlV6+LT8NDgptApQ+uIs94Q/iZAkef6sN+kKPg2DpVCtz0sl
nsMn
-----END CERTIFICATE-----`

const testKeyPEM = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDIlkHiaQ+lngtv
5bXd9I1rCqPkLyB1ksjIwi+O0eX8nfxrEoTs1SfZLbUWpEZpaKwN1u3aiDl2Yz2r
YqBGiJvgRTU+xhS7Hx3ZEgk5WhnC0y6bpV1xgIQSiKkpPANiTJupEkS/3V/RSeKg
y1R1ECxl3lduT8mO2fu+xcCpaMBABtxUvwm/Day6keE58igoL8XzfRnGvlUIBzYb
EuUJ4NnH4UobYCuWoqYhPCyjI7rMN4tcFrClS0oAfCKaopdD/NpUdfbywiiO9lel
K4BNN4UBZfNmVYQdBbo/PA8Itd5uxisaok/BI7DS7HNxrhDCli3gZOtqFHyH8wKA
50DZi4oBAgMBAAECggEABzFMH4BA7WFja5nDEbhpMcNp6Q6/jTe8N2avKeHg61QG
7xf06EEkcqcu7Sbq13DHap+gW8uys8oPNgwNTzoqzLnulNepeqQ8/8GjzwiqD9Be
xXoChcVj9v4rgq0COg4hLsjW6kJX4ztRAw8HckMoNGbqn6wAwDyUyFxy8iAtdJ46
AMuHiycqWlB1r+G3lcONjCBbnCTF+9jTlqwRbj+lLmfJrK5Qcu4ck+p2DnLN4l+w
bb71TIKFT8pVQyLM39NaqZ/RPVSgC1zb9pRnjCAouM3oxw8V4bZ6lCAaPiQL+JJK
9B9JeLT1v1eqglMBRiFiN0mqQOM0L8lkUWoUsWI8oQKBgQD+C3D7DjKoYc5FYsQb
RFo/L4tpx0AoWJ9MYxGmB19MbO+ez61Wl6kgCUimzM9b6N8xfdbQQkqDGDFX1qU+
K6ZgDIrzCTvr0lwMOg1kAAfYzHE7rNA9gSeNdmZJ0E0olvonjeRlR/qF/g8Z6oMk
9jeeZjthFLV8iBns26pWaWdQbQKBgQDKIXw1nZsEL39qIOK5gFw/Hs6jhlBoUIsN
SVz1jWQzlhmH8waW/BTPuz4JxM5Xya7CSmGIbJvr6WuTe935Kb4ht9olHyK82NPi
5px2cWKP/xKcmmE5cxYi6hhmM5WUcO61Eg2zkfOFbvbbYYdXvUO5dBMmRRb8MAsv
DHdLO1CrZQKBgQDLYHl4Yytm5bXukl0Qvy3Ie9WOPzc7lYch7gXI9wnx8xv59aR9
ODjrLsN81WYD3HAh/O4mF4vzW0DVYz8ygFtXdXMfkfrolaWfHDJwJh4iD7lu3rBv
LKBvfaPx39KFdiiZ1dxMwMzszDFmu/l1c6+fHZTX6W5JXePzXQAG4acWGQKBgAZ2
67ILSFnp6vlJ8/Za1JhwM8unEAtGCCx0nDR+QSYlNsvSSfOqPEAojONjF/ZWzPAJ
0PS8BICXBonA/GhrnVkWuDNXu5Sumpg3J+nh1nUkg0Pe7B0aQSr8sasTG0WUFw5T
dXy4vkEWO27ov5tewju8KqCetQ17u9/VVjthukLBAoGAVODrBNj0z5sTX5BlwQav
7OknKQ4vrK59BDmH81Iuua2dHsWlj2fvG0S/viJFk7B4ZvVEJYAlDjEPsS1s8D8e
BZDF+QNf8nct1iI7mdkwahTKuab3RzSzqebX63QQGi09ZVRT7QAhZfYeyP8WnRgV
stLMbpU+pSscwQn+TdLucVQ=
-----END PRIVATE KEY-----`

// ---------------------------------------------------------------------------
// Helper builders
// ---------------------------------------------------------------------------

func newBackend(name, server string, healthy bool) *backendv1.Backend {
	return &backendv1.Backend{
		Meta:   &metav1.Meta{Name: name},
		Config: &backendv1.BackendConfig{Server: server},
		Status: &backendv1.BackendStatus{Healthy: healthy},
	}
}

func newBackendWithTLS(name, server, caRef, authRef string, insecure bool, healthy bool) *backendv1.Backend {
	return &backendv1.Backend{
		Meta: &metav1.Meta{Name: name},
		Config: &backendv1.BackendConfig{
			Server:                server,
			CaRef:                 caRef,
			AuthRef:               authRef,
			InsecureSkipTlsVerify: insecure,
		},
		Status: &backendv1.BackendStatus{Healthy: healthy},
	}
}

func newRoute(name, backendRef, pathPrefix, sni string) *routev1.Route {
	return &routev1.Route{
		Meta: &metav1.Meta{Name: name},
		Config: &routev1.RouteConfig{
			BackendRef: backendRef,
			Match: &routev1.Match{
				PathPrefix: pathPrefix,
				Sni:        sni,
			},
		},
	}
}

func newRouteWithHeader(name, backendRef, pathPrefix, sni, headerName, headerValue string) *routev1.Route {
	r := newRoute(name, backendRef, pathPrefix, sni)
	r.Config.Match.Header = &routev1.HeaderMatch{
		Name:  headerName,
		Value: headerValue,
	}
	return r
}

func newRouteWithJWT(name, backendRef, pathPrefix, claim, value string) *routev1.Route {
	return &routev1.Route{
		Meta: &metav1.Meta{Name: name},
		Config: &routev1.RouteConfig{
			BackendRef: backendRef,
			Match: &routev1.Match{
				PathPrefix: pathPrefix,
				Jwt: &routev1.JWTMatch{
					Claim: claim,
					Value: value,
				},
			},
		},
	}
}

func newCA(name, certPEM string) *cav1.CertificateAuthority {
	return &cav1.CertificateAuthority{
		Meta:   &metav1.Meta{Name: name},
		Config: &cav1.CertificateAuthorityConfig{Certificate: certPEM},
	}
}

func newCert(name, certPEM, keyPEM string) *certificatev1.Certificate {
	return &certificatev1.Certificate{
		Meta: &metav1.Meta{Name: name},
		Config: &certificatev1.CertificateConfig{
			Certificate: certPEM,
			Key:         keyPEM,
		},
	}
}

// makeJWT creates a minimal unsigned JWT with the given claims map.
// It is NOT cryptographically valid, but the compiler's JWT middleware
// only checks claim presence and value, not the signature.
func makeJWT(claims map[string]interface{}) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	payloadBytes, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	return fmt.Sprintf("%s.%s.fakesig", header, payload)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestCompile_EmptyCache(t *testing.T) {
	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends:               map[string]*backendv1.Backend{},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)
	require.NotNil(t, rt)
	assert.Equal(t, uint64(1), rt.Version)
	ln := rt.Listeners["default"]
	require.NotNil(t, ln)
	assert.Empty(t, ln.ExactHosts)
	assert.Empty(t, ln.WildcardHosts)
}

func TestCompile_UnhealthyBackend_Skipped(t *testing.T) {
	// An unhealthy backend should produce no routes even if a route references it.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, false), // unhealthy
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", "/api", ""),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	ln := rt.Listeners["default"]
	require.NotNil(t, ln)
	// The route references an unhealthy backend, so the vhost should have no routes.
	assert.Empty(t, ln.ExactHosts)
}

func TestCompile_PathType_Prefix(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", "/api/v1", ""),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	ln := rt.Listeners["default"]
	vh, ok := ln.ExactHosts["*"]
	require.True(t, ok, "route with no SNI should land in ExactHosts[\"*\"]")
	assert.Len(t, vh.PrefixPaths, 1)
	assert.Equal(t, "/api/v1", vh.PrefixPaths[0].Path)
}

func TestCompile_PathType_CatchAll(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", "", ""), // no path prefix → catch-all
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	vh, ok := rt.Listeners["default"].ExactHosts["*"]
	require.True(t, ok)
	assert.Len(t, vh.CatchAll, 1)
	assert.Empty(t, vh.PrefixPaths)
	assert.Empty(t, vh.ExactPaths)
}

func TestCompile_PrefixPaths_SortedLongestFirst(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"short": newRoute("short", "be", "/api", ""),
			"long":  newRoute("long", "be", "/api/v1/resources", ""),
			"mid":   newRoute("mid", "be", "/api/v1", ""),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	vh := rt.Listeners["default"].ExactHosts["*"]
	require.Len(t, vh.PrefixPaths, 3)

	// Descending length order.
	assert.Equal(t, "/api/v1/resources", vh.PrefixPaths[0].Path)
	assert.Equal(t, "/api/v1", vh.PrefixPaths[1].Path)
	assert.Equal(t, "/api", vh.PrefixPaths[2].Path)
}

func TestCompile_HeaderMatch(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRouteWithHeader("r", "be", "/", "", "X-Tenant", "acme"),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	vh := rt.Listeners["default"].ExactHosts["*"]
	require.Len(t, vh.PrefixPaths, 1)
	route := vh.PrefixPaths[0]
	require.Len(t, route.Headers, 1)
	assert.Equal(t, "X-Tenant", route.Headers[0].Name)
	assert.Equal(t, "acme", route.Headers[0].Value)
}

func TestCompile_SNI_ExactHost(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", "/", "api.example.com"),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	ln := rt.Listeners["default"]
	_, ok := ln.ExactHosts["api.example.com"]
	assert.True(t, ok, "exact SNI should appear in ExactHosts")
	_, noWild := ln.ExactHosts["*"]
	assert.False(t, noWild)
}

func TestCompile_SNI_WildcardHost(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", "/", "*.example.com"),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	ln := rt.Listeners["default"]
	require.Len(t, ln.WildcardHosts, 1)
	assert.Equal(t, ".example.com", ln.WildcardHosts[0].Suffix)
}

func TestCompile_SNI_NoSNI_CatchAllBucket(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", "/", ""), // empty SNI
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	ln := rt.Listeners["default"]
	_, ok := ln.ExactHosts["*"]
	assert.True(t, ok, "no-SNI route should land in ExactHosts[\"*\"]")
}

func TestCompile_CA_Resolution(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	_, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackendWithTLS("be", upstream.URL, "myca", "", false, true),
		},
		Routes:       map[string]*routev1.Route{},
		Certificates: map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{
			"myca": newCA("myca", testCertPEM),
		},
	})
	require.NoError(t, err)
}

func TestCompile_CA_RefNotFound_Error(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer upstream.Close()

	c := compile.NewCompiler()
	_, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackendWithTLS("be", upstream.URL, "nonexistent-ca", "", false, true),
		},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ca_ref")
}

func TestCompile_AuthRef_Resolution(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	_, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackendWithTLS("be", upstream.URL, "", "mycert", false, true),
		},
		Routes: map[string]*routev1.Route{},
		Certificates: map[string]*certificatev1.Certificate{
			"mycert": newCert("mycert", testCertPEM, testKeyPEM),
		},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)
}

func TestCompile_AuthRef_NotFound_Error(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer upstream.Close()

	c := compile.NewCompiler()
	_, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackendWithTLS("be", upstream.URL, "", "nonexistent-cert", false, true),
		},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth_ref")
}

func TestCompile_InvalidCACert_Error(t *testing.T) {
	c := compile.NewCompiler()
	_, err := c.Compile(&compile.State{
		Backends:     map[string]*backendv1.Backend{},
		Routes:       map[string]*routev1.Route{},
		Certificates: map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{
			"bad": newCA("bad", "not-a-pem"),
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compile CAs")
}

func TestCompile_RouteWithMissingBackend_Skipped(t *testing.T) {
	// A route that references a backend that doesn't exist should be silently skipped.
	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{},
		Routes: map[string]*routev1.Route{
			"orphan": newRoute("orphan", "missing-backend", "/", ""),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)
	assert.Empty(t, rt.Listeners["default"].ExactHosts)
}

// ---------------------------------------------------------------------------
// JWT middleware integration tests
// ---------------------------------------------------------------------------

func TestJWTMiddleware_ValidClaim_Passes(t *testing.T) {
	// Set up a real upstream that returns 200.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRouteWithJWT("r", "be", "/secure", "role", "admin"),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	// Retrieve the compiled route's handler.
	vh, ok := rt.Listeners["default"].ExactHosts["*"]
	require.True(t, ok)
	require.Len(t, vh.PrefixPaths, 1)
	handler := vh.PrefixPaths[0].Handler

	token := makeJWT(map[string]interface{}{"role": "admin", "sub": "user1"})
	req := httptest.NewRequest(http.MethodGet, "/secure/resource", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTMiddleware_WrongClaimValue_Returns401(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRouteWithJWT("r", "be", "/secure", "role", "admin"),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	vh := rt.Listeners["default"].ExactHosts["*"]
	handler := vh.PrefixPaths[0].Handler

	token := makeJWT(map[string]interface{}{"role": "viewer"}) // wrong value
	req := httptest.NewRequest(http.MethodGet, "/secure/resource", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.True(t, strings.Contains(rec.Body.String(), "mismatch"))
}

func TestJWTMiddleware_MissingClaim_Returns401(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRouteWithJWT("r", "be", "/secure", "role", "admin"),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	vh := rt.Listeners["default"].ExactHosts["*"]
	handler := vh.PrefixPaths[0].Handler

	// Token does not contain the "role" claim at all.
	token := makeJWT(map[string]interface{}{"sub": "user1"})
	req := httptest.NewRequest(http.MethodGet, "/secure/resource", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.True(t, strings.Contains(rec.Body.String(), "not found"))
}

func TestJWTMiddleware_NoAuthorizationHeader_Returns401(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRouteWithJWT("r", "be", "/secure", "role", "admin"),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	vh := rt.Listeners["default"].ExactHosts["*"]
	handler := vh.PrefixPaths[0].Handler

	req := httptest.NewRequest(http.MethodGet, "/secure/resource", nil) // no auth header
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTMiddleware_MalformedToken_Returns401(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRouteWithJWT("r", "be", "/secure", "role", "admin"),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	vh := rt.Listeners["default"].ExactHosts["*"]
	handler := vh.PrefixPaths[0].Handler

	req := httptest.NewRequest(http.MethodGet, "/secure/resource", nil)
	req.Header.Set("Authorization", "Bearer notajwt")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ---------------------------------------------------------------------------
// End-to-end forwarding test
// ---------------------------------------------------------------------------

func TestCompile_E2E_RouteForwarding(t *testing.T) {
	// Real upstream that echoes back a fixed response.
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Backend", "upstream")
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	c := compile.NewCompiler()
	rt, err := c.Compile(&compile.State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", upstream.URL, true),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", "/", ""),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
	})
	require.NoError(t, err)

	vh := rt.Listeners["default"].ExactHosts["*"]
	require.Len(t, vh.PrefixPaths, 1)
	handler := vh.PrefixPaths[0].Handler

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "upstream", rec.Header().Get("X-Backend"))
}
