package config

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	types "github.com/amimof/multikube/api/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/clientcmd"
)

func TestExportKubeconfig_TokenAuth(t *testing.T) {
	caPEM := "-----BEGIN CERTIFICATE-----\nfake-ca\n-----END CERTIFICATE-----"
	caB64 := base64.StdEncoding.EncodeToString([]byte(caPEM))

	cfg := &types.Config{
		Backends: []*types.Backend{
			{
				Name:    "staging",
				Server:  "https://staging.k8s.example.com:6443",
				CaRef:   "staging-ca",
				AuthRef: "staging-cred",
			},
		},
		CertificateAuthorities: []*types.CertificateAuthority{
			{Name: "staging-ca", CertificateData: caB64},
		},
		Credentials: []*types.Credential{
			{Name: "staging-cred", Token: "my-bearer-token"},
		},
	}

	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "staging", &buf)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(buf.Bytes())
	require.NoError(t, err)

	// Cluster.
	require.Contains(t, kubecfg.Clusters, "staging")
	assert.Equal(t, "https://staging.k8s.example.com:6443", kubecfg.Clusters["staging"].Server)
	assert.Equal(t, []byte(caPEM), kubecfg.Clusters["staging"].CertificateAuthorityData)
	assert.False(t, kubecfg.Clusters["staging"].InsecureSkipTLSVerify)

	// AuthInfo.
	require.Contains(t, kubecfg.AuthInfos, "staging")
	assert.Equal(t, "my-bearer-token", kubecfg.AuthInfos["staging"].Token)

	// Context.
	require.Contains(t, kubecfg.Contexts, "staging")
	assert.Equal(t, "staging", kubecfg.Contexts["staging"].Cluster)
	assert.Equal(t, "staging", kubecfg.Contexts["staging"].AuthInfo)
	assert.Equal(t, "staging", kubecfg.CurrentContext)
}

func TestExportKubeconfig_ClientCertAuth(t *testing.T) {
	certPEM := "-----BEGIN CERTIFICATE-----\nclient-cert\n-----END CERTIFICATE-----"
	keyPEM := "-----BEGIN RSA PRIVATE KEY-----\nclient-key\n-----END RSA PRIVATE KEY-----"

	cfg := &types.Config{
		Backends: []*types.Backend{
			{
				Name:    "prod",
				Server:  "https://prod.k8s.example.com:6443",
				AuthRef: "prod-cred",
			},
		},
		Certificates: []*types.Certificate{
			{Name: "prod-cert", CertificateData: certPEM, KeyData: keyPEM},
		},
		Credentials: []*types.Credential{
			{Name: "prod-cred", ClientCertificateRef: "prod-cert"},
		},
	}

	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "prod", &buf)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(buf.Bytes())
	require.NoError(t, err)

	require.Contains(t, kubecfg.AuthInfos, "prod")
	assert.Equal(t, []byte(certPEM), kubecfg.AuthInfos["prod"].ClientCertificateData)
	assert.Equal(t, []byte(keyPEM), kubecfg.AuthInfos["prod"].ClientKeyData)
	assert.Empty(t, kubecfg.AuthInfos["prod"].Token)
}

func TestExportKubeconfig_BasicAuth(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "dev", Server: "https://dev.k8s.example.com:6443", AuthRef: "dev-cred"},
		},
		Credentials: []*types.Credential{
			{
				Name: "dev-cred",
				Basic: &types.CredentialBasic{
					Username: "admin",
					Password: "secret",
				},
			},
		},
	}

	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "dev", &buf)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(buf.Bytes())
	require.NoError(t, err)

	require.Contains(t, kubecfg.AuthInfos, "dev")
	assert.Equal(t, "admin", kubecfg.AuthInfos["dev"].Username)
	assert.Equal(t, "secret", kubecfg.AuthInfos["dev"].Password)
}

func TestExportKubeconfig_NoAuth(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{
			{
				Name:                  "insecure",
				Server:                "https://insecure.example.com:6443",
				InsecureSkipTlsVerify: true,
			},
		},
	}

	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "insecure", &buf)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(buf.Bytes())
	require.NoError(t, err)

	require.Contains(t, kubecfg.Clusters, "insecure")
	assert.True(t, kubecfg.Clusters["insecure"].InsecureSkipTLSVerify)
	require.Contains(t, kubecfg.AuthInfos, "insecure")
	// AuthInfo should be empty (no auth method).
	assert.Empty(t, kubecfg.AuthInfos["insecure"].Token)
	assert.Nil(t, kubecfg.AuthInfos["insecure"].ClientCertificateData)
}

func TestExportKubeconfig_CAFromFile(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.crt")
	caPEM := []byte("-----BEGIN CERTIFICATE-----\nfile-ca\n-----END CERTIFICATE-----")
	require.NoError(t, os.WriteFile(caPath, caPEM, 0o644))

	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "file-ca-backend", Server: "https://example.com:6443", CaRef: "my-ca"},
		},
		CertificateAuthorities: []*types.CertificateAuthority{
			{Name: "my-ca", Certificate: caPath},
		},
	}

	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "file-ca-backend", &buf)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(buf.Bytes())
	require.NoError(t, err)

	assert.Equal(t, caPEM, kubecfg.Clusters["file-ca-backend"].CertificateAuthorityData)
}

func TestExportKubeconfig_CertFromFile(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "client.crt")
	keyPath := filepath.Join(dir, "client.key")
	certPEM := []byte("-----BEGIN CERTIFICATE-----\nfile-cert\n-----END CERTIFICATE-----")
	keyPEM := []byte("-----BEGIN RSA PRIVATE KEY-----\nfile-key\n-----END RSA PRIVATE KEY-----")
	require.NoError(t, os.WriteFile(certPath, certPEM, 0o644))
	require.NoError(t, os.WriteFile(keyPath, keyPEM, 0o600))

	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "file-cert-backend", Server: "https://example.com:6443", AuthRef: "my-cred"},
		},
		Certificates: []*types.Certificate{
			{Name: "my-cert", Certificate: certPath, Key: keyPath},
		},
		Credentials: []*types.Credential{
			{Name: "my-cred", ClientCertificateRef: "my-cert"},
		},
	}

	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "file-cert-backend", &buf)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(buf.Bytes())
	require.NoError(t, err)

	assert.Equal(t, certPEM, kubecfg.AuthInfos["file-cert-backend"].ClientCertificateData)
	assert.Equal(t, keyPEM, kubecfg.AuthInfos["file-cert-backend"].ClientKeyData)
}

func TestExportKubeconfig_RawPEMCA(t *testing.T) {
	// CA stored as raw PEM (not base64-encoded).
	rawPEM := "-----BEGIN CERTIFICATE-----\nraw-pem-ca\n-----END CERTIFICATE-----"

	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "raw-ca", Server: "https://example.com:6443", CaRef: "raw-ca-entry"},
		},
		CertificateAuthorities: []*types.CertificateAuthority{
			{Name: "raw-ca-entry", CertificateData: rawPEM},
		},
	}

	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "raw-ca", &buf)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(buf.Bytes())
	require.NoError(t, err)

	// Raw PEM should be passed through as-is.
	assert.Equal(t, []byte(rawPEM), kubecfg.Clusters["raw-ca"].CertificateAuthorityData)
}

func TestExportKubeconfig_BackendNotFound(t *testing.T) {
	cfg := &types.Config{}
	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "nonexistent", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `backend "nonexistent" not found`)
}

func TestExportKubeconfig_MissingCARef(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "bad-ca", Server: "https://example.com:6443", CaRef: "missing-ca"},
		},
	}

	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "bad-ca", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `ca_ref "missing-ca" not found`)
}

func TestExportKubeconfig_MissingAuthRef(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "bad-auth", Server: "https://example.com:6443", AuthRef: "missing-cred"},
		},
	}

	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "bad-auth", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `auth_ref "missing-cred" not found`)
}

func TestExportKubeconfig_MissingCertRef(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "bad-cert", Server: "https://example.com:6443", AuthRef: "cred"},
		},
		Credentials: []*types.Credential{
			{Name: "cred", ClientCertificateRef: "missing-cert"},
		},
	}

	var buf bytes.Buffer
	err := ExportKubeconfig(cfg, "bad-cert", &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `unknown certificate "missing-cert"`)
}

// --- ExportAllKubeconfigs tests ---

func TestExportAllKubeconfigs_MultipleBackends(t *testing.T) {
	caPEM := "-----BEGIN CERTIFICATE-----\nfake-ca\n-----END CERTIFICATE-----"
	caB64 := base64.StdEncoding.EncodeToString([]byte(caPEM))

	cfg := &types.Config{
		Backends: []*types.Backend{
			{
				Name:    "staging",
				Server:  "https://staging.k8s.example.com:6443",
				CaRef:   "staging-ca",
				AuthRef: "staging-cred",
			},
			{
				Name:    "prod",
				Server:  "https://prod.k8s.example.com:6443",
				AuthRef: "prod-cred",
			},
		},
		CertificateAuthorities: []*types.CertificateAuthority{
			{Name: "staging-ca", CertificateData: caB64},
		},
		Credentials: []*types.Credential{
			{Name: "staging-cred", Token: "staging-token"},
			{Name: "prod-cred", Token: "prod-token"},
		},
	}

	var buf bytes.Buffer
	err := ExportAllKubeconfigs(cfg, &buf)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(buf.Bytes())
	require.NoError(t, err)

	// Both backends present.
	require.Contains(t, kubecfg.Clusters, "staging")
	require.Contains(t, kubecfg.Clusters, "prod")
	assert.Equal(t, "https://staging.k8s.example.com:6443", kubecfg.Clusters["staging"].Server)
	assert.Equal(t, "https://prod.k8s.example.com:6443", kubecfg.Clusters["prod"].Server)

	// CA only on staging.
	assert.Equal(t, []byte(caPEM), kubecfg.Clusters["staging"].CertificateAuthorityData)
	assert.Nil(t, kubecfg.Clusters["prod"].CertificateAuthorityData)

	// Auth.
	require.Contains(t, kubecfg.AuthInfos, "staging")
	require.Contains(t, kubecfg.AuthInfos, "prod")
	assert.Equal(t, "staging-token", kubecfg.AuthInfos["staging"].Token)
	assert.Equal(t, "prod-token", kubecfg.AuthInfos["prod"].Token)

	// Contexts.
	require.Contains(t, kubecfg.Contexts, "staging")
	require.Contains(t, kubecfg.Contexts, "prod")

	// Current context is the first backend.
	assert.Equal(t, "staging", kubecfg.CurrentContext)
}

func TestExportAllKubeconfigs_SingleBackend(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "only", Server: "https://only.example.com:6443"},
		},
	}

	var buf bytes.Buffer
	err := ExportAllKubeconfigs(cfg, &buf)
	require.NoError(t, err)

	kubecfg, err := clientcmd.Load(buf.Bytes())
	require.NoError(t, err)

	require.Len(t, kubecfg.Clusters, 1)
	require.Contains(t, kubecfg.Clusters, "only")
	assert.Equal(t, "only", kubecfg.CurrentContext)
}

func TestExportAllKubeconfigs_NoBackends(t *testing.T) {
	cfg := &types.Config{}

	var buf bytes.Buffer
	err := ExportAllKubeconfigs(cfg, &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no backends defined")
}

func TestExportAllKubeconfigs_ErrorInOneBackend(t *testing.T) {
	cfg := &types.Config{
		Backends: []*types.Backend{
			{Name: "good", Server: "https://good.example.com:6443"},
			{Name: "bad", Server: "https://bad.example.com:6443", CaRef: "missing-ca"},
		},
	}

	var buf bytes.Buffer
	err := ExportAllKubeconfigs(cfg, &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `ca_ref "missing-ca" not found`)
}
