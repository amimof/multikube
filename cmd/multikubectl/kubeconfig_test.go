package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"
)

func TestRunKubeconfig_SingleBackend(t *testing.T) {
	dir := t.TempDir()

	mkYAML := `server:
  https:
    address: "proxy.example.com:8443"
backends:
  - name: staging
    server: https://staging.k8s.local:6443
routes:
  - name: staging-route
    backend_ref: staging
    match:
      path_prefix: /staging
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	output := captureStdout(t, func() {
		err := runKubeconfig([]string{"staging"}, configPath)
		require.NoError(t, err)
	})

	kubecfg, err := clientcmd.Load(output)
	require.NoError(t, err)

	require.Contains(t, kubecfg.Clusters, "staging")
	assert.Equal(t, "https://proxy.example.com:8443/staging", kubecfg.Clusters["staging"].Server)
	assert.Equal(t, "staging", kubecfg.CurrentContext)

	// AuthInfo should be empty (proxy handles backend auth).
	require.Contains(t, kubecfg.AuthInfos, "staging")
	assert.Empty(t, kubecfg.AuthInfos["staging"].Token)
}

func TestRunKubeconfig_AllBackends(t *testing.T) {
	dir := t.TempDir()

	mkYAML := `server:
  https:
    address: "proxy.example.com:8443"
backends:
  - name: staging
    server: https://staging.k8s.local:6443
  - name: prod
    server: https://prod.k8s.local:6443
routes:
  - name: staging-route
    backend_ref: staging
    match:
      path_prefix: /stg
  - name: prod-route
    backend_ref: prod
    match:
      path_prefix: /prd
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	output := captureStdout(t, func() {
		err := runKubeconfig(nil, configPath)
		require.NoError(t, err)
	})

	kubecfg, err := clientcmd.Load(output)
	require.NoError(t, err)

	require.Contains(t, kubecfg.Clusters, "staging")
	require.Contains(t, kubecfg.Clusters, "prod")
	assert.Equal(t, "https://proxy.example.com:8443/stg", kubecfg.Clusters["staging"].Server)
	assert.Equal(t, "https://proxy.example.com:8443/prd", kubecfg.Clusters["prod"].Server)
	assert.Equal(t, "staging", kubecfg.CurrentContext)
}

func TestRunKubeconfig_WithCA(t *testing.T) {
	dir := t.TempDir()

	caPEM := "-----BEGIN CERTIFICATE-----\nfake-proxy-ca\n-----END CERTIFICATE-----"
	caB64 := base64.StdEncoding.EncodeToString([]byte(caPEM))

	mkYAML := `server:
  https:
    address: "proxy.example.com:8443"
    ca_data: ` + caB64 + `
backends:
  - name: staging
    server: https://staging.k8s.local:6443
routes:
  - name: staging-route
    backend_ref: staging
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	output := captureStdout(t, func() {
		err := runKubeconfig([]string{"staging"}, configPath)
		require.NoError(t, err)
	})

	kubecfg, err := clientcmd.Load(output)
	require.NoError(t, err)

	require.Contains(t, kubecfg.Clusters, "staging")
	assert.Equal(t, []byte(caPEM), kubecfg.Clusters["staging"].CertificateAuthorityData)
}

func TestRunKubeconfig_MissingConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "nonexistent.yaml")

	err := runKubeconfig([]string{"staging"}, configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loading multikube config")
}

func TestRunKubeconfig_BackendNotFound(t *testing.T) {
	dir := t.TempDir()

	mkYAML := `backends:
  - name: staging
    server: https://staging.k8s.local:6443
routes:
  - name: staging-route
    backend_ref: staging
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	err := runKubeconfig([]string{"nonexistent"}, configPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `backend "nonexistent" not found`)
}

func TestRunKubeconfig_NoHTTPSDefaultURL(t *testing.T) {
	dir := t.TempDir()

	mkYAML := `backends:
  - name: staging
    server: https://staging.k8s.local:6443
routes:
  - name: staging-route
    backend_ref: staging
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	output := captureStdout(t, func() {
		err := runKubeconfig([]string{"staging"}, configPath)
		require.NoError(t, err)
	})

	kubecfg, err := clientcmd.Load(output)
	require.NoError(t, err)

	require.Contains(t, kubecfg.Clusters, "staging")
	assert.Equal(t, "https://localhost:8443/staging", kubecfg.Clusters["staging"].Server)
}

func TestRunKubeconfig_DefaultPathPrefix(t *testing.T) {
	dir := t.TempDir()

	mkYAML := `server:
  https:
    address: "proxy.example.com:8443"
backends:
  - name: my-cluster
    server: https://my-cluster.k8s.local:6443
routes:
  - name: my-route
    backend_ref: my-cluster
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	output := captureStdout(t, func() {
		err := runKubeconfig([]string{"my-cluster"}, configPath)
		require.NoError(t, err)
	})

	kubecfg, err := clientcmd.Load(output)
	require.NoError(t, err)

	// Without path_prefix in the route match, defaults to /backendName.
	require.Contains(t, kubecfg.Clusters, "my-cluster")
	assert.Equal(t, "https://proxy.example.com:8443/my-cluster", kubecfg.Clusters["my-cluster"].Server)
}

func TestRunKubeconfig_SkipsBackendsWithoutRoutes(t *testing.T) {
	dir := t.TempDir()

	mkYAML := `server:
  https:
    address: "proxy.example.com:8443"
backends:
  - name: staging
    server: https://staging.k8s.local:6443
  - name: dev
    server: https://dev.k8s.local:6443
routes:
  - name: staging-route
    backend_ref: staging
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	// Capture stderr for warnings.
	origStderr := os.Stderr
	rErr, wErr, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = wErr

	output := captureStdout(t, func() {
		err := runKubeconfig(nil, configPath)
		require.NoError(t, err)
	})

	_ = wErr.Close()
	os.Stderr = origStderr

	// Read warning from stderr pipe.
	warnBuf := make([]byte, 4096)
	n, _ := rErr.Read(warnBuf)
	warnOutput := string(warnBuf[:n])

	kubecfg, err := clientcmd.Load(output)
	require.NoError(t, err)

	// Only staging should be present.
	require.Contains(t, kubecfg.Clusters, "staging")
	assert.NotContains(t, kubecfg.Clusters, "dev")

	// Warning should mention dev.
	assert.Contains(t, warnOutput, "dev")
}
