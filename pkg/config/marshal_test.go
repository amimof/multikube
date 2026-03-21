package config

import (
	"testing"

	types "github.com/amimof/multikube/api/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalYAML_RoundTrip(t *testing.T) {
	// Build a Config proto, marshal to YAML, then Load back and compare.
	original := &types.Config{
		Backends: []*types.Backend{
			{
				Name:                  "prod",
				Server:                "https://k8s.example.com:6443",
				CaRef:                 "prod-ca",
				AuthRef:               "prod-cred",
				InsecureSkipTlsVerify: true,
			},
		},
		CertificateAuthorities: []*types.CertificateAuthority{
			{
				Name:            "prod-ca",
				CertificateData: "LS0tLS1CRUdJTiBDRVJU...",
			},
		},
		Credentials: []*types.Credential{
			{
				Name:  "prod-cred",
				Token: "my-token",
			},
		},
	}

	yamlBytes, err := MarshalYAML(original)
	require.NoError(t, err)

	// Load back.
	loaded, err := Load(yamlBytes)
	require.NoError(t, err)

	// Compare key fields.
	require.Len(t, loaded.Backends, 1)
	assert.Equal(t, "prod", loaded.Backends[0].Name)
	assert.Equal(t, "https://k8s.example.com:6443", loaded.Backends[0].Server)
	assert.Equal(t, "prod-ca", loaded.Backends[0].CaRef)
	assert.Equal(t, "prod-cred", loaded.Backends[0].AuthRef)
	assert.True(t, loaded.Backends[0].InsecureSkipTlsVerify)

	require.Len(t, loaded.CertificateAuthorities, 1)
	assert.Equal(t, "prod-ca", loaded.CertificateAuthorities[0].Name)
	assert.Equal(t, "LS0tLS1CRUdJTiBDRVJU...", loaded.CertificateAuthorities[0].CertificateData)

	require.Len(t, loaded.Credentials, 1)
	assert.Equal(t, "prod-cred", loaded.Credentials[0].Name)
	assert.Equal(t, "my-token", loaded.Credentials[0].Token)
}

func TestMarshalYAML_EmptyConfig(t *testing.T) {
	cfg := &types.Config{}
	yamlBytes, err := MarshalYAML(cfg)
	require.NoError(t, err)

	// Should produce valid YAML that loads back.
	loaded, err := Load(yamlBytes)
	require.NoError(t, err)
	assert.NotNil(t, loaded)
}

func TestMarshalYAML_WithCertificates(t *testing.T) {
	original := &types.Config{
		Certificates: []*types.Certificate{
			{
				Name:            "my-cert",
				CertificateData: "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
				KeyData:         "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
			},
		},
	}

	yamlBytes, err := MarshalYAML(original)
	require.NoError(t, err)

	loaded, err := Load(yamlBytes)
	require.NoError(t, err)
	require.Len(t, loaded.Certificates, 1)
	assert.Equal(t, "my-cert", loaded.Certificates[0].Name)
	assert.Equal(t, original.Certificates[0].CertificateData, loaded.Certificates[0].CertificateData)
	assert.Equal(t, original.Certificates[0].KeyData, loaded.Certificates[0].KeyData)
}

func TestMarshalYAML_BasicCredential(t *testing.T) {
	original := &types.Config{
		Credentials: []*types.Credential{
			{
				Name: "basic-cred",
				Basic: &types.CredentialBasic{
					Username: "admin",
					Password: "secret",
				},
			},
		},
	}

	yamlBytes, err := MarshalYAML(original)
	require.NoError(t, err)

	loaded, err := Load(yamlBytes)
	require.NoError(t, err)
	require.Len(t, loaded.Credentials, 1)
	assert.Equal(t, "basic-cred", loaded.Credentials[0].Name)
	require.NotNil(t, loaded.Credentials[0].Basic)
	assert.Equal(t, "admin", loaded.Credentials[0].Basic.Username)
	assert.Equal(t, "secret", loaded.Credentials[0].Basic.Password)
}

func TestMarshalYAML_ClientCertificateRef(t *testing.T) {
	original := &types.Config{
		Credentials: []*types.Credential{
			{
				Name:                 "cert-cred",
				ClientCertificateRef: "my-cert",
			},
		},
	}

	yamlBytes, err := MarshalYAML(original)
	require.NoError(t, err)

	loaded, err := Load(yamlBytes)
	require.NoError(t, err)
	require.Len(t, loaded.Credentials, 1)
	assert.Equal(t, "my-cert", loaded.Credentials[0].ClientCertificateRef)
}
