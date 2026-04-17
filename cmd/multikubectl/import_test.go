package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestDefaultImportResourceNames(t *testing.T) {
	names := defaultImportResourceNames("prod-cluster")

	if names.Backend != "prod-cluster-backend" {
		t.Fatalf("unexpected backend name: %q", names.Backend)
	}
	if names.Credential != "prod-cluster-credential" {
		t.Fatalf("unexpected credential name: %q", names.Credential)
	}
	if names.Certificate != "prod-cluster-certificate" {
		t.Fatalf("unexpected certificate name: %q", names.Certificate)
	}
	if names.CertificateAuthority != "prod-cluster-certificate-authority" {
		t.Fatalf("unexpected certificate authority name: %q", names.CertificateAuthority)
	}
}

func TestBuildImportPlanDefaultNamesNoAuth(t *testing.T) {
	kubeconfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"prod": {Server: "https://cluster.example"},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"prod": {Cluster: "prod"},
		},
	}

	plan, err := buildImportPlan("/tmp/config", "prod", kubeconfig, importResourceNames{})
	if err != nil {
		t.Fatalf("buildImportPlan returned error: %v", err)
	}

	if got := plan.Backend.GetMeta().GetName(); got != "prod-backend" {
		t.Fatalf("unexpected backend name: %q", got)
	}
	servers := plan.Backend.GetConfig().GetServers()
	if len(servers) != 1 || servers[0] != "https://cluster.example" {
		t.Fatalf("unexpected backend servers: %#v", servers)
	}
	if got := plan.Backend.GetConfig().GetAuthRef(); got != "" {
		t.Fatalf("expected empty auth ref, got %q", got)
	}
	if plan.Credential != nil {
		t.Fatal("expected no credential to be created")
	}
	if plan.Certificate != nil {
		t.Fatal("expected no certificate to be created")
	}
	if plan.CA != nil {
		t.Fatal("expected no certificate authority to be created")
	}
}

func TestBuildImportPlanCustomNames(t *testing.T) {
	kubeconfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"prod-cluster": {
				Server:                   "https://cluster.example",
				CertificateAuthorityData: []byte("CA DATA"),
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"prod-user": {Token: "secret-token"},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"prod-cluster": {Cluster: "prod-cluster", AuthInfo: "prod-user"},
		},
	}

	plan, err := buildImportPlan("/tmp/config", "prod-cluster", kubeconfig, importResourceNames{
		Backend:              "backend-a",
		Credential:           "credential-a",
		Certificate:          "certificate-a",
		CertificateAuthority: "ca-a",
	})
	if err != nil {
		t.Fatalf("buildImportPlan returned error: %v", err)
	}

	if got := plan.Backend.GetMeta().GetName(); got != "backend-a" {
		t.Fatalf("unexpected backend name: %q", got)
	}
	if got := plan.Backend.GetConfig().GetCaRef(); got != "ca-a" {
		t.Fatalf("unexpected ca ref: %q", got)
	}
	if got := plan.Backend.GetConfig().GetAuthRef(); got != "credential-a" {
		t.Fatalf("unexpected auth ref: %q", got)
	}
	if plan.CA == nil || plan.CA.GetMeta().GetName() != "ca-a" {
		t.Fatal("expected custom certificate authority to be created")
	}
	if plan.Credential == nil || plan.Credential.GetMeta().GetName() != "credential-a" {
		t.Fatal("expected custom credential to be created")
	}
	if plan.Certificate != nil {
		t.Fatal("expected no certificate to be created for token auth")
	}
}

func TestBuildImportPlanTokenAuthFromFile(t *testing.T) {
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.txt")
	if err := os.WriteFile(tokenPath, []byte("token-from-file\n"), 0o600); err != nil {
		t.Fatalf("WriteFile token: %v", err)
	}

	kubeconfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"prod": {Server: "https://cluster.example"},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"prod-user": {TokenFile: "token.txt"},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"prod": {Cluster: "prod", AuthInfo: "prod-user"},
		},
	}

	plan, err := buildImportPlan(filepath.Join(dir, "config"), "prod", kubeconfig, importResourceNames{})
	if err != nil {
		t.Fatalf("buildImportPlan returned error: %v", err)
	}

	if plan.Credential == nil {
		t.Fatal("expected credential to be created")
	}
	if got := plan.Credential.GetConfig().GetToken(); got != "token-from-file" {
		t.Fatalf("unexpected token: %q", got)
	}
}

func TestBuildImportPlanBasicAuth(t *testing.T) {
	kubeconfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"prod": {Server: "https://cluster.example"},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"prod-user": {Username: "alice", Password: "secret"},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"prod": {Cluster: "prod", AuthInfo: "prod-user"},
		},
	}

	plan, err := buildImportPlan("/tmp/config", "prod", kubeconfig, importResourceNames{})
	if err != nil {
		t.Fatalf("buildImportPlan returned error: %v", err)
	}

	basic := plan.Credential.GetConfig().GetBasic()
	if basic == nil {
		t.Fatal("expected basic credential config")
	}
	if basic.GetUsername() != "alice" || basic.GetPassword() != "secret" {
		t.Fatalf("unexpected basic auth config: %#v", basic)
	}
}

func TestBuildImportPlanMTLSAndCAFromFiles(t *testing.T) {
	dir := t.TempDir()
	caPath := filepath.Join(dir, "ca.pem")
	certPath := filepath.Join(dir, "client.pem")
	keyPath := filepath.Join(dir, "client-key.pem")

	for _, tc := range []struct {
		path    string
		content string
	}{
		{path: caPath, content: "CA PEM"},
		{path: certPath, content: "CERT PEM"},
		{path: keyPath, content: "KEY PEM"},
	} {
		if err := os.WriteFile(tc.path, []byte(tc.content), 0o600); err != nil {
			t.Fatalf("WriteFile %s: %v", tc.path, err)
		}
	}

	kubeconfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"prod": {
				Server:               "https://cluster.example",
				CertificateAuthority: "ca.pem",
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"prod-user": {
				ClientCertificate: "client.pem",
				ClientKey:         "client-key.pem",
			},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"prod": {Cluster: "prod", AuthInfo: "prod-user"},
		},
	}

	plan, err := buildImportPlan(filepath.Join(dir, "config"), "prod", kubeconfig, importResourceNames{})
	if err != nil {
		t.Fatalf("buildImportPlan returned error: %v", err)
	}

	if plan.CA == nil {
		t.Fatal("expected certificate authority to be created")
	}
	if got := plan.CA.GetConfig().GetCertificateData(); got != "CA PEM" {
		t.Fatalf("unexpected CA pem: %q", got)
	}
	if plan.Certificate == nil {
		t.Fatal("expected certificate to be created")
	}
	if plan.Credential == nil {
		t.Fatal("expected credential to be created")
	}
	if got := plan.Credential.GetConfig().GetClientCertificateRef(); got != "prod-certificate" {
		t.Fatalf("unexpected certificate ref: %q", got)
	}
}

func TestBuildImportPlanUnsupportedAuthExec(t *testing.T) {
	kubeconfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"prod": {Server: "https://cluster.example"},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"prod-user": {Exec: &clientcmdapi.ExecConfig{Command: "aws"}},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"prod": {Cluster: "prod", AuthInfo: "prod-user"},
		},
	}

	_, err := buildImportPlan("/tmp/config", "prod", kubeconfig, importResourceNames{})
	if err == nil {
		t.Fatal("expected error for exec auth")
	}
	if !strings.Contains(err.Error(), "unsupported auth method: exec") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildImportPlanMixedAuthFails(t *testing.T) {
	kubeconfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"prod": {Server: "https://cluster.example"},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"prod-user": {Token: "secret", Username: "alice", Password: "password"},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"prod": {Cluster: "prod", AuthInfo: "prod-user"},
		},
	}

	_, err := buildImportPlan("/tmp/config", "prod", kubeconfig, importResourceNames{})
	if err == nil {
		t.Fatal("expected error for mixed auth")
	}
	if !strings.Contains(err.Error(), "multiple supported auth methods") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildImportPlanIncompleteClientCertFails(t *testing.T) {
	kubeconfig := &clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			"prod": {Server: "https://cluster.example"},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"prod-user": {ClientCertificateData: []byte("CERT ONLY")},
		},
		Contexts: map[string]*clientcmdapi.Context{
			"prod": {Cluster: "prod", AuthInfo: "prod-user"},
		},
	}

	_, err := buildImportPlan("/tmp/config", "prod", kubeconfig, importResourceNames{})
	if err == nil {
		t.Fatal("expected error for incomplete client cert auth")
	}
	if !strings.Contains(err.Error(), "requires both certificate and key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewConfigCmdDoesNotExposeImport(t *testing.T) {
	cmd := newConfigCmd()
	for _, sub := range cmd.Commands() {
		if sub.Name() == "import" {
			t.Fatal("config command should not expose import")
		}
	}
}
