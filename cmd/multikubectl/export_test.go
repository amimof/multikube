package main

import (
	"bytes"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"testing"

	mkconfig "github.com/amimof/multikube/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// writeMultikubeConfig writes a multikube config YAML string to a temp file
// and returns the path.
func writeMultikubeConfig(t *testing.T, dir, yaml string) string {
	t.Helper()
	path := filepath.Join(dir, "multikube-config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(yaml), 0o644))
	return path
}

// captureStdout runs fn while capturing os.Stdout and returns the output.
func captureStdout(t *testing.T, fn func()) []byte {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := os.Stdout
	os.Stdout = w

	fn()

	defer func() {
		_ = w.Close()
	}()
	os.Stdout = orig

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)

	return buf.Bytes()
}

// --- runExport end-to-end tests ---

func TestRunExport_TokenAuth(t *testing.T) {
	dir := t.TempDir()

	caPEM := "-----BEGIN CERTIFICATE-----\nfake-ca\n-----END CERTIFICATE-----"
	caB64 := base64.StdEncoding.EncodeToString([]byte(caPEM))

	mkYAML := `backends:
  - name: staging
    server: https://staging.k8s.example.com:6443
    ca_ref: staging-ca
    auth_ref: staging-cred
certificate_authorities:
  - name: staging-ca
    certificate_data: ` + caB64 + `
credentials:
  - name: staging-cred
    token: my-bearer-token
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	output := captureStdout(t, func() {
		err := runExport("staging", configPath)
		require.NoError(t, err)
	})

	// Parse the output as kubeconfig.
	kubecfg, err := clientcmd.Load(output)
	require.NoError(t, err)

	// Cluster.
	require.Contains(t, kubecfg.Clusters, "staging")
	assert.Equal(t, "https://staging.k8s.example.com:6443", kubecfg.Clusters["staging"].Server)
	assert.Equal(t, []byte(caPEM), kubecfg.Clusters["staging"].CertificateAuthorityData)

	// AuthInfo.
	require.Contains(t, kubecfg.AuthInfos, "staging")
	assert.Equal(t, "my-bearer-token", kubecfg.AuthInfos["staging"].Token)

	// Context.
	require.Contains(t, kubecfg.Contexts, "staging")
	assert.Equal(t, "staging", kubecfg.Contexts["staging"].Cluster)
	assert.Equal(t, "staging", kubecfg.Contexts["staging"].AuthInfo)
	assert.Equal(t, "staging", kubecfg.CurrentContext)
}

func TestRunExport_ClientCertAuth(t *testing.T) {
	dir := t.TempDir()

	// Use single-line PEM stubs to avoid YAML multiline issues.
	certPEM := "-----BEGIN CERTIFICATE-----client-cert-----END CERTIFICATE-----"
	keyPEM := "-----BEGIN RSA PRIVATE KEY-----client-key-----END RSA PRIVATE KEY-----"

	mkYAML := `backends:
  - name: prod
    server: https://prod.k8s.example.com:6443
    auth_ref: prod-cred
certificates:
  - name: prod-cert
    certificate_data: "` + certPEM + `"
    key_data: "` + keyPEM + `"
credentials:
  - name: prod-cred
    client_certificate_ref: prod-cert
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	output := captureStdout(t, func() {
		err := runExport("prod", configPath)
		require.NoError(t, err)
	})

	kubecfg, err := clientcmd.Load(output)
	require.NoError(t, err)

	require.Contains(t, kubecfg.AuthInfos, "prod")
	assert.Equal(t, []byte(certPEM), kubecfg.AuthInfos["prod"].ClientCertificateData)
	assert.Equal(t, []byte(keyPEM), kubecfg.AuthInfos["prod"].ClientKeyData)
}

func TestRunExport_BackendNotFound(t *testing.T) {
	dir := t.TempDir()

	mkYAML := `backends:
  - name: prod
    server: https://prod.example.com:6443
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	err := runExport("nonexistent", configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `backend "nonexistent" not found`)
}

func TestRunExport_MissingConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "nonexistent.yaml")

	err := runExport("staging", configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loading multikube config")
}

// --- runExportAll tests ---

func TestRunExportAll_MultipleBackends(t *testing.T) {
	dir := t.TempDir()

	mkYAML := `backends:
  - name: staging
    server: https://staging.k8s.example.com:6443
    auth_ref: staging-cred
  - name: prod
    server: https://prod.k8s.example.com:6443
    auth_ref: prod-cred
credentials:
  - name: staging-cred
    token: staging-token
  - name: prod-cred
    token: prod-token
`
	configPath := writeMultikubeConfig(t, dir, mkYAML)

	output := captureStdout(t, func() {
		err := runExportAll(configPath)
		require.NoError(t, err)
	})

	kubecfg, err := clientcmd.Load(output)
	require.NoError(t, err)

	// Both backends present.
	require.Contains(t, kubecfg.Clusters, "staging")
	require.Contains(t, kubecfg.Clusters, "prod")
	assert.Equal(t, "https://staging.k8s.example.com:6443", kubecfg.Clusters["staging"].Server)
	assert.Equal(t, "https://prod.k8s.example.com:6443", kubecfg.Clusters["prod"].Server)

	// Auth.
	require.Contains(t, kubecfg.AuthInfos, "staging")
	require.Contains(t, kubecfg.AuthInfos, "prod")
	assert.Equal(t, "staging-token", kubecfg.AuthInfos["staging"].Token)
	assert.Equal(t, "prod-token", kubecfg.AuthInfos["prod"].Token)

	// Contexts.
	require.Contains(t, kubecfg.Contexts, "staging")
	require.Contains(t, kubecfg.Contexts, "prod")

	// First backend is current-context.
	assert.Equal(t, "staging", kubecfg.CurrentContext)
}

func TestRunExportAll_MissingConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "nonexistent.yaml")

	err := runExportAll(configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loading multikube config")
}

// --- Round-trip: import then export ---

func TestRoundTrip_ImportThenExport(t *testing.T) {
	dir := t.TempDir()

	// Start with a kubeconfig that has one context.
	caPEM := "-----BEGIN CERTIFICATE-----\nfake-ca\n-----END CERTIFICATE-----"

	originalKubecfg := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"myctx": {
				Server:                   "https://k8s.example.com:6443",
				CertificateAuthorityData: []byte(caPEM),
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"myctx": {Token: "original-token"},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"myctx": {Cluster: "myctx", AuthInfo: "myctx"},
		},
	}

	// Write the original kubeconfig using clientcmd.WriteToFile.
	originalKubeconfigPath := filepath.Join(dir, "original-kubeconfig")
	require.NoError(t, clientcmd.WriteToFile(*originalKubecfg, originalKubeconfigPath))

	// Import into multikube config.
	configPath := filepath.Join(dir, "multikube.yaml")
	err := runImport("myctx", originalKubeconfigPath, configPath)
	require.NoError(t, err)

	// Verify multikube config.
	cfg, err := mkconfig.LoadFromFile(configPath)
	require.NoError(t, err)
	require.Len(t, cfg.Backends, 1)
	assert.Equal(t, "myctx", cfg.Backends[0].Name)

	// Export back to stdout.
	output := captureStdout(t, func() {
		err := runExport("myctx", configPath)
		require.NoError(t, err)
	})

	// Verify the exported kubeconfig matches the original.
	exported, err := clientcmd.Load(output)
	require.NoError(t, err)

	require.Contains(t, exported.Clusters, "myctx")
	assert.Equal(t, "https://k8s.example.com:6443", exported.Clusters["myctx"].Server)
	assert.Equal(t, []byte(caPEM), exported.Clusters["myctx"].CertificateAuthorityData)

	require.Contains(t, exported.AuthInfos, "myctx")
	assert.Equal(t, "original-token", exported.AuthInfos["myctx"].Token)

	assert.Equal(t, "myctx", exported.CurrentContext)
}
