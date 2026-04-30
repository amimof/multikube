package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/amimof/multikube/pkg/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new multikube configuration file",
		Long: `Create a new, empty multikube configuration file at the default or
specified path.

If a configuration file already exists at the target path the command exits
with an error and leaves the existing file untouched.

The generated file contains the required version field and an empty list of
servers, ready to be populated with backends via 'multikubectl config import'.`,
		Example: `  # Create the configuration file at the default location (~/.config/multikubectl.yaml)
  multikubectl config init

  # Create the configuration file at a custom path
  multikubectl config init --config /etc/multikube/config.yaml`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInitCmd()
		},
	}
	return cmd
}

// runInitCmd creates an empty multikube configuration file at the path
// resolved by viper, then immediately reads it back to verify it is valid.
func runInitCmd() error {
	configPath := viper.ConfigFileUsed()

	if configExists(configPath) {
		logrus.Fatalf("config already exists at %s", configPath)
	}

	cfg := &client.Config{
		Version: "config/v1",
		Servers: []*client.Server{},
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		logrus.Fatalf("error marshal: %v", err)
	}

	err = os.MkdirAll(path.Dir(configPath), 0o755)
	if err != nil {
		logrus.Fatalf("error creating config dir: %v", err)
	}

	err = os.WriteFile(configPath, b, 0o666)
	if err != nil {
		logrus.Fatalf("error writing config file: %v", err)
	}

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("error reading config: %v", err)
	}

	fmt.Printf("Configuration created in %s\n", configPath)

	return nil
}

func configExists(p string) bool {
	_, err := os.Stat(p)
	return !errors.Is(err, os.ErrNotExist)
}
