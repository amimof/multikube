package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	authv1 "github.com/amimof/multikube/api/auth/v1"
	clientpkg "github.com/amimof/multikube/pkg/client"
	authclientv1 "github.com/amimof/multikube/pkg/client/auth/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/protobuf/proto"
)

type stubAuthClient struct {
	loginResp *authv1.LoginResponse
	loginReq  *authv1.LoginRequest
	err       error
}

type stubLoginClientSet struct {
	authClient authclientv1.ClientV1
}

func (s *stubLoginClientSet) AuthV1() authclientv1.ClientV1 {
	return s.authClient
}

func (s *stubLoginClientSet) Close() error {
	return nil
}

func (s *stubAuthClient) Login(_ context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	s.loginReq = proto.Clone(req).(*authv1.LoginRequest)
	return s.loginResp, s.err
}

func (s *stubAuthClient) Logout(context.Context, *authv1.LogoutRequest) error {
	return nil
}

func (s *stubAuthClient) Refresh(context.Context, *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	return nil, nil
}

func TestRunLoginCmdStoresTokensAndPrintsConfirmation(t *testing.T) {
	clearLoginEnv(t)
	authClient, cmd, configPath := newLoginTestCommand(t, clientpkg.Config{
		Version: "config/v1",
		Current: "prod",
		Servers: []*clientpkg.Server{{
			Name:    "prod",
			Address: "example.com:443",
			TLSConfig: &clientpkg.TLSConfig{
				Insecure: true,
			},
		}},
	})

	if err := runLoginCmd(cmd, nil, loginOptions{username: "alice", password: "super-secret"}); err != nil {
		t.Fatalf("runLoginCmd returned error: %v", err)
	}

	assertLoginRequest(t, authClient, "alice", "super-secret")

	server, err := cfg.CurrentServer()
	if err != nil {
		t.Fatalf("CurrentServer returned error: %v", err)
	}
	assertStoredSession(t, server, "access-token", "refresh-token")
	assertPersistedSession(t, configPath)

	stdout := cmd.OutOrStdout().(*bytes.Buffer).String()
	stderr := cmd.ErrOrStderr().(*bytes.Buffer).String()
	if !strings.Contains(stdout, "Logged in to prod as alice") {
		t.Fatalf("stdout = %q, want login confirmation", stdout)
	}
	if strings.Contains(stdout, "access-token") || strings.Contains(stdout, "refresh-token") {
		t.Fatalf("stdout should not contain token values: %q", stdout)
	}
	if strings.Contains(stderr, "access-token") || strings.Contains(stderr, "refresh-token") {
		t.Fatalf("stderr should not contain token values: %q", stderr)
	}
}

func TestRunLoginCmdReadsCredentialsFromEnv(t *testing.T) {
	clearLoginEnv(t)
	t.Setenv(loginUsernameEnvVar, "env-user")
	t.Setenv(loginPasswordEnvVar, "env-password")

	authClient, cmd, _ := newLoginTestCommand(t, clientpkg.Config{
		Version: "config/v1",
		Current: "prod",
		Servers: []*clientpkg.Server{{
			Name:      "prod",
			Address:   "example.com:443",
			TLSConfig: &clientpkg.TLSConfig{Insecure: true},
		}},
	})

	if err := runLoginCmd(cmd, nil, loginOptions{}); err != nil {
		t.Fatalf("runLoginCmd returned error: %v", err)
	}

	assertLoginRequest(t, authClient, "env-user", "env-password")
}

func TestRunLoginCmdFlagsOverrideEnv(t *testing.T) {
	clearLoginEnv(t)
	t.Setenv(loginUsernameEnvVar, "env-user")
	t.Setenv(loginPasswordEnvVar, "env-password")

	authClient, cmd, _ := newLoginTestCommand(t, clientpkg.Config{
		Version: "config/v1",
		Current: "prod",
		Servers: []*clientpkg.Server{{
			Name:      "prod",
			Address:   "example.com:443",
			TLSConfig: &clientpkg.TLSConfig{Insecure: true},
		}},
	})

	if err := runLoginCmd(cmd, nil, loginOptions{username: "flag-user", password: "flag-password"}); err != nil {
		t.Fatalf("runLoginCmd returned error: %v", err)
	}

	assertLoginRequest(t, authClient, "flag-user", "flag-password")
}

func TestRunLoginCmdConfiguresNewServerAndSetsCurrentByDefault(t *testing.T) {
	clearLoginEnv(t)
	authClient, cmd, _ := newLoginTestCommand(t, clientpkg.Config{
		Version: "config/v1",
		Servers: []*clientpkg.Server{},
	})

	if err := runLoginCmd(cmd, []string{"prod"}, loginOptions{
		address:  "example.com:443",
		current:  true,
		insecure: true,
		username: "alice",
		password: "super-secret",
	}); err != nil {
		t.Fatalf("runLoginCmd returned error: %v", err)
	}

	assertLoginRequest(t, authClient, "alice", "super-secret")
	if cfg.Current != "prod" {
		t.Fatalf("current = %q, want %q", cfg.Current, "prod")
	}
	server, err := cfg.GetServer("prod")
	if err != nil {
		t.Fatalf("GetServer returned error: %v", err)
	}
	if server.Address != "example.com:443" {
		t.Fatalf("address = %q, want %q", server.Address, "example.com:443")
	}
	if server.TLSConfig == nil || !server.TLSConfig.Insecure {
		t.Fatalf("tls config = %#v, want insecure true", server.TLSConfig)
	}
	assertStoredSession(t, server, "access-token", "refresh-token")
}

func TestRunLoginCmdUpdatesExistingServerInPlace(t *testing.T) {
	clearLoginEnv(t)
	_, cmd, _ := newLoginTestCommand(t, clientpkg.Config{
		Version: "config/v1",
		Current: "prod",
		Servers: []*clientpkg.Server{{
			Name:    "prod",
			Address: "old.example.com:443",
			TLSConfig: &clientpkg.TLSConfig{
				Insecure: true,
			},
		}},
	})

	if err := runLoginCmd(cmd, []string{"prod"}, loginOptions{
		address:  "new.example.com:443",
		current:  true,
		insecure: true,
		username: "alice",
		password: "super-secret",
	}); err != nil {
		t.Fatalf("runLoginCmd returned error: %v", err)
	}

	server, err := cfg.GetServer("prod")
	if err != nil {
		t.Fatalf("GetServer returned error: %v", err)
	}
	if server.Address != "new.example.com:443" {
		t.Fatalf("address = %q, want %q", server.Address, "new.example.com:443")
	}
}

func TestRunLoginCmdDoesNotPersistConfigOnLoginFailure(t *testing.T) {
	clearLoginEnv(t)
	authClient, cmd, configPath := newLoginTestCommand(t, clientpkg.Config{
		Version: "config/v1",
		Servers: []*clientpkg.Server{},
	})
	authClient.err = errors.New("boom")

	err := runLoginCmd(cmd, []string{"prod"}, loginOptions{
		address:  "example.com:443",
		current:  true,
		insecure: true,
		username: "alice",
		password: "super-secret",
	})
	if err == nil {
		t.Fatal("expected login to fail")
	}
	if len(cfg.Servers) != 0 {
		t.Fatalf("servers = %#v, want no persisted servers", cfg.Servers)
	}
	if _, statErr := os.Stat(configPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected config file to remain absent, stat err = %v", statErr)
	}
}

func TestRunLoginCmdLoadsTLSFilesForProvisionedServer(t *testing.T) {
	clearLoginEnv(t)
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

	_, cmd, _ := newLoginTestCommand(t, clientpkg.Config{Version: "config/v1", Servers: []*clientpkg.Server{}})

	if err := runLoginCmd(cmd, []string{"prod"}, loginOptions{
		address:  "example.com:443",
		current:  true,
		caFile:   caPath,
		certFile: certPath,
		keyFile:  keyPath,
		username: "alice",
		password: "super-secret",
	}); err != nil {
		t.Fatalf("runLoginCmd returned error: %v", err)
	}

	server, err := cfg.GetServer("prod")
	if err != nil {
		t.Fatalf("GetServer returned error: %v", err)
	}
	if server.TLSConfig == nil {
		t.Fatal("expected tls config to be set")
	}
	if server.TLSConfig.CA != "CA PEM" {
		t.Fatalf("ca = %q, want %q", server.TLSConfig.CA, "CA PEM")
	}
	if server.TLSConfig.Certificate != "CERT PEM" {
		t.Fatalf("certificate = %q, want %q", server.TLSConfig.Certificate, "CERT PEM")
	}
	if server.TLSConfig.Key != "KEY PEM" {
		t.Fatalf("key = %q, want %q", server.TLSConfig.Key, "KEY PEM")
	}
}

func newLoginTestCommand(t *testing.T, initialCfg clientpkg.Config) (*stubAuthClient, *cobra.Command, string) {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "multikube.yaml")
	viper.SetConfigFile(configPath)

	cfg = initialCfg

	authClient := &stubAuthClient{loginResp: &authv1.LoginResponse{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}}

	originalFactory := loginClientFactory
	t.Cleanup(func() {
		loginClientFactory = originalFactory
	})
	loginClientFactory = func(server *clientpkg.Server) (loginClient, error) {
		return &stubLoginClientSet{authClient: authClient}, nil
	}

	cmd := newLoginCmd()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetContext(context.Background())

	return authClient, cmd, configPath
}

func assertLoginRequest(t *testing.T, authClient *stubAuthClient, username, password string) {
	t.Helper()
	if authClient.loginReq == nil {
		t.Fatal("expected login request to be sent")
	}
	if authClient.loginReq.GetUsername() != username {
		t.Fatalf("username = %q, want %q", authClient.loginReq.GetUsername(), username)
	}
	if authClient.loginReq.GetPassword() != password {
		t.Fatalf("password = %q, want %q", authClient.loginReq.GetPassword(), password)
	}
}

func assertStoredSession(t *testing.T, server *clientpkg.Server, accessToken, refreshToken string) {
	t.Helper()
	if server == nil || server.Session == nil {
		t.Fatal("expected session to be stored")
	}
	if server.Session.AccessToken != accessToken {
		t.Fatalf("access token = %q, want %q", server.Session.AccessToken, accessToken)
	}
	if server.Session.RefreshToken != refreshToken {
		t.Fatalf("refresh token = %q, want %q", server.Session.RefreshToken, refreshToken)
	}
}

func assertPersistedSession(t *testing.T, configPath string) {
	t.Helper()
	written, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	contents := string(written)
	if !strings.Contains(contents, "access_token: access-token") {
		t.Fatalf("config missing access token: %s", contents)
	}
	if !strings.Contains(contents, "refresh_token: refresh-token") {
		t.Fatalf("config missing refresh token: %s", contents)
	}
}

func clearLoginEnv(t *testing.T) {
	t.Helper()
	t.Setenv(loginUsernameEnvVar, "")
	t.Setenv(loginPasswordEnvVar, "")
}
