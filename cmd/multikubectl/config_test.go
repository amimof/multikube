package main

import (
	"os"
	"path/filepath"
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
