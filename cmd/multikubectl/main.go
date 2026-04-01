package main

import (
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

var cfg client.Config

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")
}

func withConfig(run func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := loadConfig(); err != nil {
			return err
		}
		return run(cmd, args)
	}
}

func loadConfig() error {
	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatalf("error reading config: %v", err)
		return err
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		logrus.Fatalf("error decoding config into struct: %v", err)
		return err
	}
	if err := cfg.Validate(); err != nil {
		logrus.Fatalf("config validation error: %v", err)
		return err
	}
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
	defaultConfigPath := filepath.Join(home, ".multikube", "multikube.yaml")

	// Setup flags
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", defaultConfigPath, "config file")
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "localhost:5700", "Address of the API Server")
	rootCmd.PersistentFlags().StringVar(&tlsCACert, "tls-ca-certificate", "", "CA Certificate file path")
	rootCmd.PersistentFlags().StringVar(&tlsCert, "tls-certificate", "", "Certificate file path")
	rootCmd.PersistentFlags().StringVar(&tlsCertKey, "tls-certificate-key", "", "Certificate key file path")
	rootCmd.PersistentFlags().StringVar(&otelEndpoint, "otel-endpoint", "", "Endpoint address of OpenTelemetry collector")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", false, "Skip TLS certificate verification")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "number for the log level verbosity (debug, info, warn, error, fatal, panic)")

	rootCmd.AddCommand(newGetCmd(&cfg))
	rootCmd.AddCommand(newCreateCmd(&cfg))
	rootCmd.AddCommand(newDeleteCmd(&cfg))
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newKubeconfigCmd())

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
