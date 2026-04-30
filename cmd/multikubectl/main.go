package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/amimof/multikube/pkg/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// VERSION of the app. Set at build time via -ldflags.
	VERSION string
	// COMMIT is the Git commit. Set at build time via -ldflags.
	COMMIT string

	configFile   string
	logLevel     string
	server       string
	insecure     bool
	tlsCACert    string
	tlsCert      string
	tlsCertKey   string
	otelEndpoint string
	outputFormat string
	rootCmd      = cobra.Command{
		Use:   "multikubectl",
		Short: "CLI for managing multikube configuration",
		Long:  "multikubectl is a command-line tool for managing multikube configuration files.",
	}
)

var (
	cfg       client.Config
	clientSet *client.ClientSet
)

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
}

func withConfigWithoutValidation(run func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(false); err != nil {
			return err
		}
		return run(cmd, args)
	}
}

func withConfig(run func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(true); err != nil {
			return err
		}
		return run(cmd, args)
	}
}

func withClientSet(run func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return withConfig(func(cmd *cobra.Command, args []string) error {
		currentSrv, err := cfg.CurrentServer()
		if err != nil {
			logrus.Fatal(err)
		}
		clientSet, err = client.New(currentSrv.Address, client.WithTLSConfigFromCfg(&cfg))
		if err != nil {
			logrus.Fatalf("error setting up client: %v", err)
			return err
		}
		defer func() {
			if err := clientSet.Close(); err != nil {
				logrus.Fatalf("error closing client connection: %v", err)
			}
		}()
		return run(cmd, args)
	})
}

func loadConfig(validate bool) error {
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) || errors.Is(err, os.ErrNotExist) {
			cfg = client.Config{
				Version: "config/v1",
				Servers: []*client.Server{},
			}
			if validate {
				if err := cfg.Validate(); err != nil {
					logrus.Fatalf("config validation error: %v", err)
					return err
				}
			}
			return nil
		}
		logrus.Fatalf("error reading config: %v", err)
		return err
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		logrus.Fatalf("error decoding config into struct: %v", err)
		return err
	}
	if validate {
		if err := cfg.Validate(); err != nil {
			logrus.Fatalf("config validation error: %v", err)
			return err
		}
	}
	return nil
}

func buildServerConfig(serverName, address, caFile, certFile, keyFile string, insecure, tls bool) (*client.Server, error) {
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
				return nil, fmt.Errorf("error reading ca file: %v", err)
			}
			newServer.TLSConfig.CA = string(caData)
		}

		if certFile != "" {
			certData, err := os.ReadFile(certFile)
			if err != nil {
				return nil, fmt.Errorf("error reading certificate file: %v", err)
			}
			newServer.TLSConfig.Certificate = string(certData)
		}

		if keyFile != "" {
			keyData, err := os.ReadFile(keyFile)
			if err != nil {
				return nil, fmt.Errorf("error reading key file: %v", err)
			}
			newServer.TLSConfig.Key = string(keyData)
		}
	}

	if err := newServer.Validate(); err != nil {
		return nil, err
	}

	return newServer, nil
}

func upsertServerConfig(cfg *client.Config, srv *client.Server) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if srv == nil {
		return fmt.Errorf("server is nil")
	}

	if err := srv.Validate(); err != nil {
		return err
	}

	for i, existing := range cfg.Servers {
		if existing.Name == srv.Name {
			cfg.Servers[i] = srv
			return nil
		}
	}

	cfg.Servers = append(cfg.Servers, srv)
	return nil
}

func SetVersionInfo(version, commit, date, branch, goversion string) {
	rootCmd.Version = fmt.Sprintf("Version:\t%s\nCommit:\t%v\nBuilt:\t%s\nBranch:\t%s\nGo Version:\t%s\n", version, commit, date, branch, goversion)
}

func main() {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			return err
		}
		logrus.SetLevel(lvl)
		return nil
	}

	// Figure out path to default config file
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("home directory cannot be determined: %v", err)
	}
	defaultConfigPath := filepath.Join(home, ".config", "multikubectl.yaml")

	// Setup flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", defaultConfigPath, "config file")
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "localhost:5700", "Address of the API Server")
	rootCmd.PersistentFlags().StringVar(&tlsCACert, "tls-ca-certificate", "", "CA Certificate file path")
	rootCmd.PersistentFlags().StringVar(&tlsCert, "tls-certificate", "", "Certificate file path")
	rootCmd.PersistentFlags().StringVar(&tlsCertKey, "tls-certificate-key", "", "Certificate key file path")
	rootCmd.PersistentFlags().StringVar(&otelEndpoint, "otel-endpoint", "", "Endpoint address of OpenTelemetry collector")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "number for the log level verbosity (debug, info, warn, error, fatal, panic)")

	rootCmd.AddCommand(newGetCmd())
	rootCmd.AddCommand(newCreateCmd(&cfg))
	rootCmd.AddCommand(newDeleteCmd(&cfg))
	rootCmd.AddCommand(newEditCmd())
	rootCmd.AddCommand(newImportCmd(&cfg))
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newKubeconfigCmd())
	rootCmd.AddCommand(newApplyCmd())
	rootCmd.AddCommand(newLoginCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\nCommit: %s\n", VERSION, COMMIT)
		},
	}
}
