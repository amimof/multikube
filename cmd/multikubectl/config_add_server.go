package main

import (
	"fmt"
	"os"

	"github.com/amimof/multikube/pkg/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func newConfigCreateServerCmd() *cobra.Command {
	var (
		tls      bool
		insecure bool
		current  bool
		caFile   string
		certFile string
		keyFile  string
		address  string
	)
	cmd := &cobra.Command{
		Use:   "create-server NAME",
		Short: "Add a server to multikubectl client configuration",
		Long:  "Add a server to multikubectl client configuration",
		Example: `
# Create a server with TLS
multikubectl config create-server dev --address localhost:5743 --ca ca.crt
`,
		Args: cobra.ExactArgs(1),
		RunE: withConfigWithoutValidation(func(cmd *cobra.Command, args []string) error {
			return runConfigCreateServerCmd(args[0], address, caFile, certFile, keyFile, insecure, current, tls)
		}),
	}
	cmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "Skip TLS verification. Not recommended")
	cmd.PersistentFlags().BoolVar(&current, "current", true, "Set as current server")
	cmd.PersistentFlags().BoolVar(&tls, "tls", false, "Use TLS for this server")
	cmd.PersistentFlags().StringVar(&address, "address", "", "Endpoint address of the server")
	cmd.PersistentFlags().StringVar(&caFile, "ca", "", "Path to ca certificate file")
	cmd.PersistentFlags().StringVar(&certFile, "certificate", "", "Path to certificate file")
	cmd.PersistentFlags().StringVar(&keyFile, "key", "", "Path to private key file")
	return cmd
}

// runInitCmd creates an empty multikube configuration file at the path
// resolved by viper, then immediately reads it back to verify it is valid.
func runConfigCreateServerCmd(serverName, address, caFile, certFile, keyFile string, insecure, current, tls bool) error {
	newServer := &client.Server{
		Name:    serverName,
		Address: address,
	}

	if tls {
		newServer.TLSConfig = &client.TLSConfig{
			Insecure: insecure,
		}
		if caFile != "" {
			caData, err := os.ReadFile(caFile)
			if err != nil {
				return fmt.Errorf("error reading ca file: %v", err)
			}
			newServer.TLSConfig.CA = string(caData)
		}

		if certFile != "" {
			certData, err := os.ReadFile(certFile)
			if err != nil {
				return fmt.Errorf("error reading certificate file: %v", err)
			}
			newServer.TLSConfig.Certificate = string(certData)
		}

		if keyFile != "" {
			keyData, err := os.ReadFile(keyFile)
			if err != nil {
				return fmt.Errorf("error reading key file: %v", err)
			}
			newServer.TLSConfig.Key = string(keyData)
		}
	}

	if current {
		cfg.Current = newServer.Name
	}

	err := cfg.AddServer(newServer)
	if err != nil {
		return fmt.Errorf("error addding server to config: %v", err)
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshal: %v", err)
	}

	err = os.WriteFile(viper.GetViper().ConfigFileUsed(), b, 0o666)
	if err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	logrus.Infof("Added server %v to configuration", serverName)

	return nil
}
