package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	clientpkg "github.com/amimof/multikube/pkg/client"
	"github.com/spf13/viper"
)

func TestLoadConfigMissingFileInitializesEmptyConfigWithoutValidation(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "missing", "multikubectl.yaml")
	viper.SetConfigFile(configPath)

	if err := loadConfig(false); err != nil {
		t.Fatalf("loadConfig returned error: %v", err)
	}

	if cfg.Version != "config/v1" {
		t.Fatalf("version = %q, want %q", cfg.Version, "config/v1")
	}
	if len(cfg.Servers) != 0 {
		t.Fatalf("servers = %#v, want empty", cfg.Servers)
	}
}

func TestWriteConfigCreatesParentDirectory(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "nested", "path", "multikubectl.yaml")
	viper.SetConfigFile(configPath)
	cfg = clientpkg.Config{
		Version: "config/v1",
		Current: "prod",
		Servers: []*clientpkg.Server{{
			Name:    "prod",
			Address: "example.com:443",
			TLSConfig: &clientpkg.TLSConfig{
				Insecure: true,
			},
		}},
	}

	if err := writeConfig(); err != nil {
		t.Fatalf("writeConfig returned error: %v", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Stat returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(configPath)); err != nil {
		t.Fatalf("parent dir Stat returned error: %v", err)
	}
}

func TestPersistCurrentSessionTokenUpdatesConfigAndWritesFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "multikubectl.yaml")
	viper.SetConfigFile(configPath)
	cfg = clientpkg.Config{
		Version: "config/v1",
		Current: "prod",
		Servers: []*clientpkg.Server{{
			Name:    "prod",
			Address: "example.com:443",
			TLSConfig: &clientpkg.TLSConfig{
				Insecure: true,
			},
		}},
	}

	err := persistCurrentSessionToken(&clientpkg.Token{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	})
	if err != nil {
		t.Fatalf("persistCurrentSessionToken returned error: %v", err)
	}

	server, err := cfg.CurrentServer()
	if err != nil {
		t.Fatalf("CurrentServer returned error: %v", err)
	}
	if server.Session == nil {
		t.Fatal("expected session to be set")
	}
	if server.Session.AccessToken != "new-access-token" {
		t.Fatalf("access token = %q, want %q", server.Session.AccessToken, "new-access-token")
	}
	if server.Session.RefreshToken != "new-refresh-token" {
		t.Fatalf("refresh token = %q, want %q", server.Session.RefreshToken, "new-refresh-token")
	}

	b, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	contents := string(b)
	if !strings.Contains(contents, "access_token: new-access-token") {
		t.Fatalf("config missing access token: %s", contents)
	}
	if !strings.Contains(contents, "refresh_token: new-refresh-token") {
		t.Fatalf("config missing refresh token: %s", contents)
	}
}
