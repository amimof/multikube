package config

import (
	"bytes"
	"encoding/base64"
	"testing"

	types "github.com/amimof/multikube/api/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"
)

func TestGenerateProxyKubeconfig_SingleBackend(t *testing.T) {
	cfg := &types.Config{
		Server: &types.Server{
			Https: &types.HTTPSListener{Address: "proxy.example.com:8443"},
		},
		Backends: []*types.Backend{
			{Name: "staging", Server: "https://staging.k8s.local:6443"},
		},
		Routes: []*types.Route{
			{Name: "staging-route", BackendRef: "staging", Match: &types.Match{PathPrefix: "/staging"}},
		},
	}

	var out bytes.Buffer
	err := GenerateProxyKubeconfig(cfg, []string{"staging"}, &out, nil)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(out.Bytes())
	require.NoError(t, err)

	require.Contains(t, kubecfg.Clusters, "staging")
	assert.Equal(t, "https://proxy.example.com:8443/staging", kubecfg.Clusters["staging"].Server)
	assert.Equal(t, "staging", kubecfg.CurrentContext)
	require.Contains(t, kubecfg.Contexts, "staging")
	assert.Equal(t, "staging", kubecfg.Contexts["staging"].Cluster)
	assert.Equal(t, "staging", kubecfg.Contexts["staging"].AuthInfo)
	// AuthInfo should be empty (no upstream creds).
	require.Contains(t, kubecfg.AuthInfos, "staging")
	assert.Empty(t, kubecfg.AuthInfos["staging"].Token)
}

func TestGenerateProxyKubeconfig_DefaultPathPrefix(t *testing.T) {
	cfg := &types.Config{
		Server: &types.Server{
			Https: &types.HTTPSListener{Address: "proxy.example.com:8443"},
		},
		Backends: []*types.Backend{
			{Name: "prod", Server: "https://prod.k8s.local:6443"},
		},
		Routes: []*types.Route{
			{Name: "prod-route", BackendRef: "prod"},
		},
	}

	var out bytes.Buffer
	err := GenerateProxyKubeconfig(cfg, []string{"prod"}, &out, nil)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(out.Bytes())
	require.NoError(t, err)

	// Without PathPrefix in route, falls back to /<backendName>.
	require.Contains(t, kubecfg.Clusters, "prod")
	assert.Equal(t, "https://proxy.example.com:8443/prod", kubecfg.Clusters["prod"].Server)
}

func TestGenerateProxyKubeconfig_AllBackendsWithRoutes(t *testing.T) {
	cfg := &types.Config{
		Server: &types.Server{
			Https: &types.HTTPSListener{Address: "proxy.example.com:8443"},
		},
		Backends: []*types.Backend{
			{Name: "staging", Server: "https://staging.k8s.local:6443"},
			{Name: "prod", Server: "https://prod.k8s.local:6443"},
		},
		Routes: []*types.Route{
			{Name: "staging-route", BackendRef: "staging", Match: &types.Match{PathPrefix: "/stg"}},
			{Name: "prod-route", BackendRef: "prod", Match: &types.Match{PathPrefix: "/prd"}},
		},
	}

	var out bytes.Buffer
	err := GenerateProxyKubeconfig(cfg, nil, &out, nil)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(out.Bytes())
	require.NoError(t, err)

	require.Contains(t, kubecfg.Clusters, "staging")
	require.Contains(t, kubecfg.Clusters, "prod")
	assert.Equal(t, "https://proxy.example.com:8443/stg", kubecfg.Clusters["staging"].Server)
	assert.Equal(t, "https://proxy.example.com:8443/prd", kubecfg.Clusters["prod"].Server)
	assert.Equal(t, "staging", kubecfg.CurrentContext)
}

func TestGenerateProxyKubeconfig_HTTPSWithCA(t *testing.T) {
	caPEM := "-----BEGIN CERTIFICATE-----\nfake-proxy-ca\n-----END CERTIFICATE-----"
	caB64 := base64.StdEncoding.EncodeToString([]byte(caPEM))

	cfg := &types.Config{
		Server: &types.Server{
			Https: &types.HTTPSListener{
				Address:  "proxy.example.com:8443",
				CaSource: &types.HTTPSListener_CaData{CaData: caB64},
			},
		},
		Backends: []*types.Backend{
			{Name: "staging", Server: "https://staging.k8s.local:6443"},
		},
		Routes: []*types.Route{
			{Name: "staging-route", BackendRef: "staging"},
		},
	}

	var out bytes.Buffer
	err := GenerateProxyKubeconfig(cfg, []string{"staging"}, &out, nil)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(out.Bytes())
	require.NoError(t, err)

	require.Contains(t, kubecfg.Clusters, "staging")
	assert.Equal(t, []byte(caPEM), kubecfg.Clusters["staging"].CertificateAuthorityData)
}

func TestGenerateProxyKubeconfig_NoHTTPSDefault(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "staging", Server: "https://staging.k8s.local:6443"},
		},
		Routes: []*types.Route{
			{Name: "staging-route", BackendRef: "staging"},
		},
	}

	var out bytes.Buffer
	err := GenerateProxyKubeconfig(cfg, []string{"staging"}, &out, nil)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(out.Bytes())
	require.NoError(t, err)

	require.Contains(t, kubecfg.Clusters, "staging")
	assert.Equal(t, "https://localhost:8443/staging", kubecfg.Clusters["staging"].Server)
}

func TestGenerateProxyKubeconfig_UnixOnlyError(t *testing.T) {
	cfg := &types.Config{
		Server: &types.Server{
			Unix: &types.UnixListener{Path: "/var/run/multikube.sock"},
		},
		Backends: []*types.Backend{
			{Name: "staging", Server: "https://staging.k8s.local:6443"},
		},
		Routes: []*types.Route{
			{Name: "staging-route", BackendRef: "staging"},
		},
	}

	var out bytes.Buffer
	err := GenerateProxyKubeconfig(cfg, []string{"staging"}, &out, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unix")
}

func TestGenerateProxyKubeconfig_NoRouteSkip(t *testing.T) {
	cfg := &types.Config{
		Server: &types.Server{
			Https: &types.HTTPSListener{Address: "proxy.example.com:8443"},
		},
		Backends: []*types.Backend{
			{Name: "staging", Server: "https://staging.k8s.local:6443"},
			{Name: "dev", Server: "https://dev.k8s.local:6443"},
		},
		Routes: []*types.Route{
			{Name: "staging-route", BackendRef: "staging"},
		},
	}

	var out bytes.Buffer
	var warn bytes.Buffer
	err := GenerateProxyKubeconfig(cfg, nil, &out, &warn)
	require.NoError(t, err)

	// Only staging should be in the output; dev has no route.
	kubecfg, err := clientcmd.Load(out.Bytes())
	require.NoError(t, err)

	require.Contains(t, kubecfg.Clusters, "staging")
	assert.NotContains(t, kubecfg.Clusters, "dev")

	// Warning should mention dev.
	assert.Contains(t, warn.String(), "dev")
}

func TestGenerateProxyKubeconfig_ExplicitBackendNoRouteError(t *testing.T) {
	cfg := &types.Config{
		Server: &types.Server{
			Https: &types.HTTPSListener{Address: "proxy.example.com:8443"},
		},
		Backends: []*types.Backend{
			{Name: "staging", Server: "https://staging.k8s.local:6443"},
		},
		Routes: []*types.Route{},
	}

	var out bytes.Buffer
	err := GenerateProxyKubeconfig(cfg, []string{"staging"}, &out, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no route found")
}

func TestGenerateProxyKubeconfig_AllBackendsNoRoutes(t *testing.T) {
	cfg := &types.Config{
		Server: &types.Server{
			Https: &types.HTTPSListener{Address: "proxy.example.com:8443"},
		},
		Backends: []*types.Backend{
			{Name: "staging", Server: "https://staging.k8s.local:6443"},
		},
		Routes: []*types.Route{},
	}

	var out bytes.Buffer
	var warn bytes.Buffer
	err := GenerateProxyKubeconfig(cfg, nil, &out, &warn)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no backends with routes found")
}

func TestGenerateProxyKubeconfig_BackendNotFound(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{},
	}

	var out bytes.Buffer
	err := GenerateProxyKubeconfig(cfg, []string{"nonexistent"}, &out, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `backend "nonexistent" not found`)
}

func TestGenerateProxyKubeconfig_WildcardAddressSubstitution(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{"bare port", ":8443"},
		{"0.0.0.0", "0.0.0.0:8443"},
		{":: (IPv6 wildcard)", "[::]:8443"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &types.Config{
				Server: &types.Server{
					Https: &types.HTTPSListener{Address: tt.address},
				},
				Backends: []*types.Backend{
					{Name: "staging", Server: "https://staging.k8s.local:6443"},
				},
				Routes: []*types.Route{
					{Name: "staging-route", BackendRef: "staging"},
				},
			}

			var out bytes.Buffer
			err := GenerateProxyKubeconfig(cfg, []string{"staging"}, &out, nil)
			require.NoError(t, err)

			kubecfg, err := clientcmd.Load(out.Bytes())
			require.NoError(t, err)

			assert.Equal(t, "https://localhost:8443/staging", kubecfg.Clusters["staging"].Server)
		})
	}
}

// --- resolveListenerURL unit tests ---

func TestResolveListenerURL_NoServer(t *testing.T) {
	cfg := &types.Config{}
	url, err := resolveListenerURL(cfg)
	require.NoError(t, err)
	assert.Equal(t, "https://localhost:8443", url)
}

func TestResolveListenerURL_EmptyServer(t *testing.T) {
	cfg := &types.Config{Server: &types.Server{}}
	url, err := resolveListenerURL(cfg)
	require.NoError(t, err)
	assert.Equal(t, "https://localhost:8443", url)
}

func TestResolveListenerURL_WithAddress(t *testing.T) {
	cfg := &types.Config{
		Server: &types.Server{
			Https: &types.HTTPSListener{Address: "myhost.example.com:9443"},
		},
	}
	url, err := resolveListenerURL(cfg)
	require.NoError(t, err)
	assert.Equal(t, "https://myhost.example.com:9443", url)
}

func TestResolveListenerURL_BarePort(t *testing.T) {
	cfg := &types.Config{
		Server: &types.Server{
			Https: &types.HTTPSListener{Address: ":8443"},
		},
	}
	url, err := resolveListenerURL(cfg)
	require.NoError(t, err)
	assert.Equal(t, "https://localhost:8443", url)
}

// --- findRouteForBackend unit tests ---

func TestFindRouteForBackend_Found(t *testing.T) {
	cfg := &types.Config{
		Routes: []*types.Route{
			{Name: "a", BackendRef: "alpha"},
			{Name: "b", BackendRef: "beta"},
		},
	}
	r := findRouteForBackend(cfg, "beta")
	require.NotNil(t, r)
	assert.Equal(t, "b", r.Name)
}

func TestFindRouteForBackend_NotFound(t *testing.T) {
	cfg := &types.Config{
		Routes: []*types.Route{
			{Name: "a", BackendRef: "alpha"},
		},
	}
	r := findRouteForBackend(cfg, "gamma")
	assert.Nil(t, r)
}

// --- buildPathPrefix unit tests ---

func TestBuildPathPrefix_ExplicitPrefix(t *testing.T) {
	route := &types.Route{
		Match: &types.Match{PathPrefix: "/custom-prefix"},
	}
	assert.Equal(t, "/custom-prefix", buildPathPrefix(route, "mybackend"))
}

func TestBuildPathPrefix_ExplicitPrefixNoLeadingSlash(t *testing.T) {
	route := &types.Route{
		Match: &types.Match{PathPrefix: "custom-prefix"},
	}
	assert.Equal(t, "/custom-prefix", buildPathPrefix(route, "mybackend"))
}

func TestBuildPathPrefix_FallbackToBackendName(t *testing.T) {
	route := &types.Route{}
	assert.Equal(t, "/mybackend", buildPathPrefix(route, "mybackend"))
}

func TestBuildPathPrefix_EmptyMatchPathPrefix(t *testing.T) {
	route := &types.Route{
		Match: &types.Match{PathPrefix: ""},
	}
	assert.Equal(t, "/mybackend", buildPathPrefix(route, "mybackend"))
}
