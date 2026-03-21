package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	types "github.com/amimof/multikube/api/config/v1"
	mkconfig "github.com/amimof/multikube/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd/api"
)

// --- buildImportObjects tests ---

func TestBuildImportObjects_FullContext(t *testing.T) {
	cluster := &api.Cluster{
		Server:                   "https://k8s.prod.example.com:6443",
		CertificateAuthorityData: []byte("-----BEGIN CERTIFICATE-----\nfake-ca\n-----END CERTIFICATE-----"),
		InsecureSkipTLSVerify:    false,
	}
	authInfo := &api.AuthInfo{
		ClientCertificateData: []byte("-----BEGIN CERTIFICATE-----\nfake-cert\n-----END CERTIFICATE-----"),
		ClientKeyData:         []byte("-----BEGIN RSA PRIVATE KEY-----\nfake-key\n-----END RSA PRIVATE KEY-----"),
	}

	result, err := buildImportObjects("prod", cluster, authInfo)
	require.NoError(t, err)

	// Backend.
	assert.Equal(t, "prod", result.Backend.Name)
	assert.Equal(t, "https://k8s.prod.example.com:6443", result.Backend.Server)
	assert.Equal(t, "prod-ca", result.Backend.CaRef)
	assert.Equal(t, "prod-cred", result.Backend.AuthRef)
	assert.False(t, result.Backend.InsecureSkipTlsVerify)

	// CA — stored as base64 of the PEM bytes.
	require.NotNil(t, result.CertificateAuthority)
	assert.Equal(t, "prod-ca", result.CertificateAuthority.Name)
	decoded, err := base64.StdEncoding.DecodeString(result.CertificateAuthority.CertificateData)
	require.NoError(t, err)
	assert.Equal(t, cluster.CertificateAuthorityData, decoded)

	// Client certificate — stored as raw PEM string.
	require.NotNil(t, result.Certificate)
	assert.Equal(t, "prod-cert", result.Certificate.Name)
	assert.Equal(t, string(authInfo.ClientCertificateData), result.Certificate.CertificateData)
	assert.Equal(t, string(authInfo.ClientKeyData), result.Certificate.KeyData)

	// Credential — references the client certificate.
	require.NotNil(t, result.Credential)
	assert.Equal(t, "prod-cred", result.Credential.Name)
	assert.Equal(t, "prod-cert", result.Credential.ClientCertificateRef)
}

func TestBuildImportObjects_BearerToken(t *testing.T) {
	cluster := &api.Cluster{
		Server: "https://k8s.staging.example.com:6443",
	}
	authInfo := &api.AuthInfo{
		Token: "my-bearer-token-123",
	}

	result, err := buildImportObjects("staging", cluster, authInfo)
	require.NoError(t, err)

	assert.Equal(t, "staging", result.Backend.Name)
	assert.Nil(t, result.CertificateAuthority) // no CA
	assert.Nil(t, result.Certificate)          // no client cert

	require.NotNil(t, result.Credential)
	assert.Equal(t, "staging-cred", result.Credential.Name)
	assert.Equal(t, "my-bearer-token-123", result.Credential.Token)
	assert.Empty(t, result.Backend.CaRef)
	assert.Equal(t, "staging-cred", result.Backend.AuthRef)
}

func TestBuildImportObjects_BasicAuth(t *testing.T) {
	cluster := &api.Cluster{
		Server: "https://k8s.dev.example.com:6443",
	}
	authInfo := &api.AuthInfo{
		Username: "admin",
		Password: "p@ssw0rd",
	}

	result, err := buildImportObjects("dev", cluster, authInfo)
	require.NoError(t, err)

	require.NotNil(t, result.Credential)
	assert.Equal(t, "dev-cred", result.Credential.Name)
	require.NotNil(t, result.Credential.Basic)
	assert.Equal(t, "admin", result.Credential.Basic.Username)
	assert.Equal(t, "p@ssw0rd", result.Credential.Basic.Password)
}

func TestBuildImportObjects_NoAuth(t *testing.T) {
	cluster := &api.Cluster{
		Server:                "https://k8s.example.com:6443",
		InsecureSkipTLSVerify: true,
	}

	result, err := buildImportObjects("insecure", cluster, nil)
	require.NoError(t, err)

	assert.Equal(t, "insecure", result.Backend.Name)
	assert.True(t, result.Backend.InsecureSkipTlsVerify)
	assert.Nil(t, result.CertificateAuthority)
	assert.Nil(t, result.Certificate)
	assert.Nil(t, result.Credential)
	assert.Empty(t, result.Backend.AuthRef)
}

func TestBuildImportObjects_CAFilePath(t *testing.T) {
	cluster := &api.Cluster{
		Server:               "https://k8s.example.com:6443",
		CertificateAuthority: "/etc/kubernetes/pki/ca.crt",
	}

	result, err := buildImportObjects("file-ca", cluster, nil)
	require.NoError(t, err)

	require.NotNil(t, result.CertificateAuthority)
	assert.Equal(t, "file-ca-ca", result.CertificateAuthority.Name)
	assert.Equal(t, "/etc/kubernetes/pki/ca.crt", result.CertificateAuthority.Certificate)
	assert.Empty(t, result.CertificateAuthority.CertificateData)
}

func TestBuildImportObjects_CertFilePaths(t *testing.T) {
	cluster := &api.Cluster{
		Server: "https://k8s.example.com:6443",
	}
	authInfo := &api.AuthInfo{
		ClientCertificate: "/home/user/.certs/client.crt",
		ClientKey:         "/home/user/.certs/client.key",
	}

	result, err := buildImportObjects("file-cert", cluster, authInfo)
	require.NoError(t, err)

	require.NotNil(t, result.Certificate)
	assert.Equal(t, "file-cert-cert", result.Certificate.Name)
	assert.Equal(t, "/home/user/.certs/client.crt", result.Certificate.Certificate)
	assert.Equal(t, "/home/user/.certs/client.key", result.Certificate.Key)

	require.NotNil(t, result.Credential)
	assert.Equal(t, "file-cert-cert", result.Credential.ClientCertificateRef)
}

func TestBuildImportObjects_NoServer(t *testing.T) {
	cluster := &api.Cluster{
		Server: "",
	}

	_, err := buildImportObjects("empty", cluster, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no server URL")
}

// --- checkNameConflicts tests ---

func TestCheckNameConflicts_NoConflict(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "existing"},
		},
	}
	r := &importResult{
		Backend: &types.Backend{Name: "new-backend"},
	}
	assert.NoError(t, checkNameConflicts(cfg, r))
}

func TestCheckNameConflicts_BackendConflict(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "prod"},
		},
	}
	r := &importResult{
		Backend: &types.Backend{Name: "prod"},
	}
	err := checkNameConflicts(cfg, r)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `backend "prod" already exists`)
}

func TestCheckNameConflicts_CAConflict(t *testing.T) {
	cfg := &types.Config{
		CertificateAuthorities: []*types.CertificateAuthority{
			{Name: "prod-ca"},
		},
	}
	r := &importResult{
		Backend:              &types.Backend{Name: "prod"},
		CertificateAuthority: &types.CertificateAuthority{Name: "prod-ca"},
	}
	err := checkNameConflicts(cfg, r)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `certificate authority "prod-ca" already exists`)
}

func TestCheckNameConflicts_CertConflict(t *testing.T) {
	cfg := &types.Config{
		Certificates: []*types.Certificate{
			{Name: "prod-cert"},
		},
	}
	r := &importResult{
		Backend:     &types.Backend{Name: "prod"},
		Certificate: &types.Certificate{Name: "prod-cert"},
	}
	err := checkNameConflicts(cfg, r)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `certificate "prod-cert" already exists`)
}

func TestCheckNameConflicts_CredentialConflict(t *testing.T) {
	cfg := &types.Config{
		Credentials: []*types.Credential{
			{Name: "prod-cred"},
		},
	}
	r := &importResult{
		Backend:    &types.Backend{Name: "prod"},
		Credential: &types.Credential{Name: "prod-cred"},
	}
	err := checkNameConflicts(cfg, r)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `credential "prod-cred" already exists`)
}

// --- loadOrCreateConfig tests ---

func TestLoadOrCreateConfig_FileNotExist(t *testing.T) {
	cfg, err := loadOrCreateConfig("/tmp/nonexistent-multikube-config-test.yaml")
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Backends)
}

func TestLoadOrCreateConfig_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	yamlContent := `backends:
  - name: existing
    server: https://existing.example.com:6443
`
	require.NoError(t, os.WriteFile(path, []byte(yamlContent), 0o644))

	cfg, err := loadOrCreateConfig(path)
	require.NoError(t, err)
	require.Len(t, cfg.Backends, 1)
	assert.Equal(t, "existing", cfg.Backends[0].Name)
}

// --- End-to-end runImport tests ---

func writeKubeconfig(t *testing.T, dir string, kubecfg *api.Config) string {
	t.Helper()
	// Use clientcmd to write a kubeconfig file.
	path := filepath.Join(dir, "kubeconfig")

	// Manually write a minimal kubeconfig YAML.
	// clientcmd.WriteToFile requires a full config with proper types,
	// so we'll use the clientcmd package.
	err := writeKubeconfigFile(path, kubecfg)
	require.NoError(t, err)
	return path
}

// writeKubeconfigFile writes a kubeconfig api.Config to a file using
// the clientcmd serialization.
func writeKubeconfigFile(path string, config *api.Config) error {
	// Use clientcmd's built-in writer. It expects the config to have
	// a proper APIVersion and Kind set for serialization.
	return writeKubeconfigManual(path, config)
}

// writeKubeconfigManual writes a kubeconfig manually as YAML.
func writeKubeconfigManual(path string, config *api.Config) error {
	// Build a minimal kubeconfig YAML from the api.Config.
	// This is simpler than pulling in the full clientcmd serializer.
	content := "apiVersion: v1\nkind: Config\nclusters:\n"

	for name, cluster := range config.Clusters {
		content += "- name: " + name + "\n"
		content += "  cluster:\n"
		content += "    server: " + cluster.Server + "\n"
		if len(cluster.CertificateAuthorityData) > 0 {
			encoded := base64.StdEncoding.EncodeToString(cluster.CertificateAuthorityData)
			content += "    certificate-authority-data: " + encoded + "\n"
		}
		if cluster.CertificateAuthority != "" {
			content += "    certificate-authority: " + cluster.CertificateAuthority + "\n"
		}
		if cluster.InsecureSkipTLSVerify {
			content += "    insecure-skip-tls-verify: true\n"
		}
	}

	content += "users:\n"
	for name, authInfo := range config.AuthInfos {
		content += "- name: " + name + "\n"
		content += "  user:\n"
		if len(authInfo.ClientCertificateData) > 0 {
			encoded := base64.StdEncoding.EncodeToString(authInfo.ClientCertificateData)
			content += "    client-certificate-data: " + encoded + "\n"
		}
		if len(authInfo.ClientKeyData) > 0 {
			encoded := base64.StdEncoding.EncodeToString(authInfo.ClientKeyData)
			content += "    client-key-data: " + encoded + "\n"
		}
		if authInfo.Token != "" {
			content += "    token: " + authInfo.Token + "\n"
		}
		if authInfo.Username != "" {
			content += "    username: " + authInfo.Username + "\n"
		}
		if authInfo.Password != "" {
			content += "    password: " + authInfo.Password + "\n"
		}
	}

	content += "contexts:\n"
	for name, ctx := range config.Contexts {
		content += "- name: " + name + "\n"
		content += "  context:\n"
		content += "    cluster: " + ctx.Cluster + "\n"
		content += "    user: " + ctx.AuthInfo + "\n"
	}

	content += "current-context: \"\"\n"

	return os.WriteFile(path, []byte(content), 0o644)
}

func TestRunImport_EndToEnd_TokenAuth(t *testing.T) {
	dir := t.TempDir()

	// Create a kubeconfig with a token-authenticated context.
	kubecfg := &api.Config{
		Clusters: map[string]*api.Cluster{
			"staging-cluster": {
				Server:                   "https://staging.k8s.example.com:6443",
				CertificateAuthorityData: []byte("-----BEGIN CERTIFICATE-----\nfake-ca-pem\n-----END CERTIFICATE-----"),
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"staging-user": {
				Token: "eyJhbGciOiJSUzI1NiJ9.fake-token",
			},
		},
		Contexts: map[string]*api.Context{
			"staging": {
				Cluster:  "staging-cluster",
				AuthInfo: "staging-user",
			},
		},
	}

	kubeconfigPath := writeKubeconfig(t, dir, kubecfg)
	configPath := filepath.Join(dir, "multikube-config.yaml")

	err := runImport("staging", kubeconfigPath, configPath)
	require.NoError(t, err)

	// Verify the config file was written.
	cfg, err := mkconfig.LoadFromFile(configPath)
	require.NoError(t, err)

	// Backend.
	require.Len(t, cfg.Backends, 1)
	assert.Equal(t, "staging", cfg.Backends[0].Name)
	assert.Equal(t, "https://staging.k8s.example.com:6443", cfg.Backends[0].Server)
	assert.Equal(t, "staging-ca", cfg.Backends[0].CaRef)
	assert.Equal(t, "staging-cred", cfg.Backends[0].AuthRef)

	// CA.
	require.Len(t, cfg.CertificateAuthorities, 1)
	assert.Equal(t, "staging-ca", cfg.CertificateAuthorities[0].Name)
	assert.NotEmpty(t, cfg.CertificateAuthorities[0].CertificateData)

	// Credential (token).
	require.Len(t, cfg.Credentials, 1)
	assert.Equal(t, "staging-cred", cfg.Credentials[0].Name)
	assert.Equal(t, "eyJhbGciOiJSUzI1NiJ9.fake-token", cfg.Credentials[0].Token)
}

func TestRunImport_EndToEnd_ClientCert(t *testing.T) {
	dir := t.TempDir()

	kubecfg := &api.Config{
		Clusters: map[string]*api.Cluster{
			"prod-cluster": {
				Server: "https://prod.k8s.example.com:6443",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"prod-user": {
				ClientCertificateData: []byte("-----BEGIN CERTIFICATE-----\nclient-cert\n-----END CERTIFICATE-----"),
				ClientKeyData:         []byte("-----BEGIN RSA PRIVATE KEY-----\nclient-key\n-----END RSA PRIVATE KEY-----"),
			},
		},
		Contexts: map[string]*api.Context{
			"prod": {
				Cluster:  "prod-cluster",
				AuthInfo: "prod-user",
			},
		},
	}

	kubeconfigPath := writeKubeconfig(t, dir, kubecfg)
	configPath := filepath.Join(dir, "multikube-config.yaml")

	err := runImport("prod", kubeconfigPath, configPath)
	require.NoError(t, err)

	cfg, err := mkconfig.LoadFromFile(configPath)
	require.NoError(t, err)

	// Client certificate.
	require.Len(t, cfg.Certificates, 1)
	assert.Equal(t, "prod-cert", cfg.Certificates[0].Name)
	assert.Contains(t, cfg.Certificates[0].CertificateData, "client-cert")
	assert.Contains(t, cfg.Certificates[0].KeyData, "client-key")

	// Credential references the certificate.
	require.Len(t, cfg.Credentials, 1)
	assert.Equal(t, "prod-cert", cfg.Credentials[0].ClientCertificateRef)
}

func TestRunImport_EndToEnd_AppendToExisting(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Write an existing config file with one backend.
	existingYAML := `backends:
  - name: existing
    server: https://existing.example.com:6443
`
	require.NoError(t, os.WriteFile(configPath, []byte(existingYAML), 0o644))

	kubecfg := &api.Config{
		Clusters: map[string]*api.Cluster{
			"new-cluster": {
				Server: "https://new.k8s.example.com:6443",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			"new-user": {
				Token: "new-token",
			},
		},
		Contexts: map[string]*api.Context{
			"new": {
				Cluster:  "new-cluster",
				AuthInfo: "new-user",
			},
		},
	}

	kubeconfigPath := writeKubeconfig(t, dir, kubecfg)

	err := runImport("new", kubeconfigPath, configPath)
	require.NoError(t, err)

	// Verify both backends exist.
	cfg, err := mkconfig.LoadFromFile(configPath)
	require.NoError(t, err)
	require.Len(t, cfg.Backends, 2)

	names := []string{cfg.Backends[0].Name, cfg.Backends[1].Name}
	assert.Contains(t, names, "existing")
	assert.Contains(t, names, "new")
}

func TestRunImport_ContextNotFound(t *testing.T) {
	dir := t.TempDir()

	kubecfg := &api.Config{
		Clusters:  map[string]*api.Cluster{},
		AuthInfos: map[string]*api.AuthInfo{},
		Contexts:  map[string]*api.Context{},
	}

	kubeconfigPath := writeKubeconfig(t, dir, kubecfg)
	configPath := filepath.Join(dir, "config.yaml")

	err := runImport("nonexistent", kubeconfigPath, configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `context "nonexistent" not found`)
}

func TestRunImport_DuplicateBackend(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	// Existing config already has a "prod" backend.
	existingYAML := `backends:
  - name: prod
    server: https://old-prod.example.com:6443
`
	require.NoError(t, os.WriteFile(configPath, []byte(existingYAML), 0o644))

	kubecfg := &api.Config{
		Clusters: map[string]*api.Cluster{
			"prod-cluster": {
				Server: "https://new-prod.k8s.example.com:6443",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{},
		Contexts: map[string]*api.Context{
			"prod": {
				Cluster:  "prod-cluster",
				AuthInfo: "",
			},
		},
	}

	kubeconfigPath := writeKubeconfig(t, dir, kubecfg)

	err := runImport("prod", kubeconfigPath, configPath)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `backend "prod" already exists`)
}

func TestRunImport_CreatesNewFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "subdir", "config.yaml")

	kubecfg := &api.Config{
		Clusters: map[string]*api.Cluster{
			"demo-cluster": {
				Server: "https://demo.k8s.example.com:6443",
			},
		},
		AuthInfos: map[string]*api.AuthInfo{},
		Contexts: map[string]*api.Context{
			"demo": {
				Cluster:  "demo-cluster",
				AuthInfo: "",
			},
		},
	}

	kubeconfigPath := writeKubeconfig(t, dir, kubecfg)

	err := runImport("demo", kubeconfigPath, configPath)
	require.NoError(t, err)

	// Verify file was created in the subdirectory.
	_, err = os.Stat(configPath)
	assert.NoError(t, err)

	cfg, err := mkconfig.LoadFromFile(configPath)
	require.NoError(t, err)
	require.Len(t, cfg.Backends, 1)
	assert.Equal(t, "demo", cfg.Backends[0].Name)
}

func TestRunImport_InsecureSkipTLS(t *testing.T) {
	dir := t.TempDir()

	kubecfg := &api.Config{
		Clusters: map[string]*api.Cluster{
			"insecure-cluster": {
				Server:                "https://insecure.k8s.example.com:6443",
				InsecureSkipTLSVerify: true,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{},
		Contexts: map[string]*api.Context{
			"insecure": {
				Cluster:  "insecure-cluster",
				AuthInfo: "",
			},
		},
	}

	kubeconfigPath := writeKubeconfig(t, dir, kubecfg)
	configPath := filepath.Join(dir, "config.yaml")

	err := runImport("insecure", kubeconfigPath, configPath)
	require.NoError(t, err)

	cfg, err := mkconfig.LoadFromFile(configPath)
	require.NoError(t, err)
	require.Len(t, cfg.Backends, 1)
	assert.True(t, cfg.Backends[0].InsecureSkipTlsVerify)
}
