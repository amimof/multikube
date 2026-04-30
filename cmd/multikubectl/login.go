package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	authv1 "github.com/amimof/multikube/api/auth/v1"
	"github.com/amimof/multikube/pkg/client"
	authclientv1 "github.com/amimof/multikube/pkg/client/auth/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"gopkg.in/yaml.v2"
)

var passwordReader = readPassword

const (
	loginUsernameEnvVar = "MULTIKUBECTL_USERNAME"
	loginPasswordEnvVar = "MULTIKUBECTL_PASSWORD"
)

func newLoginCmd() *cobra.Command {
	var (
		address  string
		caFile   string
		certFile string
		keyFile  string
		current  bool
		insecure bool
		username string
		password string
	)

	cmd := &cobra.Command{
		Use:   "login [NAME]",
		Short: "Authenticate with the current multikube server",
		Long: "Authenticate with the current multikube server and persist the session in the CLI config. " +
			"Credentials are resolved from flags, then " + loginUsernameEnvVar + "/" + loginPasswordEnvVar + ", then interactive prompts.",
		Args: cobra.MaximumNArgs(1),
		RunE: withConfigWithoutValidation(func(cmd *cobra.Command, args []string) error {
			return runLoginCmd(cmd, args, loginOptions{
				address:  address,
				caFile:   caFile,
				certFile: certFile,
				keyFile:  keyFile,
				current:  current,
				insecure: insecure,
				username: username,
				password: password,
			})
		}),
	}

	cmd.Flags().StringVar(&address, "address", "", "Endpoint address of the server")
	cmd.Flags().StringVar(&caFile, "ca", "", "Path to ca certificate file")
	cmd.Flags().StringVar(&certFile, "certificate", "", "Path to certificate file")
	cmd.Flags().StringVar(&keyFile, "key", "", "Path to private key file")
	cmd.Flags().BoolVar(&current, "current", true, "Set the configured server as current")
	cmd.Flags().BoolVar(&insecure, "insecure", false, "Skip TLS verification. Not recommended")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Username to authenticate with; defaults to "+loginUsernameEnvVar)
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password to authenticate with; defaults to "+loginPasswordEnvVar)

	return cmd
}

type loginOptions struct {
	address  string
	caFile   string
	certFile string
	keyFile  string
	current  bool
	insecure bool
	username string
	password string
}

type loginClient interface {
	AuthV1() authclientv1.ClientV1
	Close() error
}

func runLoginCmd(cmd *cobra.Command, args []string, opts loginOptions) error {
	targetServer, shouldUpdateCurrent, err := resolveLoginServer(args, opts)
	if err != nil {
		return err
	}

	username := opts.username
	password := opts.password

	if username == "" {
		username = strings.TrimSpace(os.Getenv(loginUsernameEnvVar))
	}

	if username == "" {
		username, err = readLine(cmd.ErrOrStderr(), os.Stdin, "Username: ")
		if err != nil {
			return err
		}
	}

	if password == "" {
		password = strings.TrimSpace(os.Getenv(loginPasswordEnvVar))
	}

	if password == "" {
		password, err = passwordReader(cmd.ErrOrStderr(), "Password: ")
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	loginClient, err := loginClientFactory(targetServer)
	if err != nil {
		return err
	}
	defer func() {
		if err := loginClient.Close(); err != nil {
			_ = err
		}
	}()

	resp, err := loginClient.AuthV1().Login(ctx, &authv1.LoginRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return err
	}

	if err := upsertServerConfig(&cfg, targetServer); err != nil {
		return err
	}
	if shouldUpdateCurrent {
		cfg.Current = targetServer.Name
	}

	persistedServer, err := cfg.GetServer(targetServer.Name)
	if err != nil {
		return err
	}

	if persistedServer.Session == nil {
		persistedServer.Session = &client.Session{}
	}
	persistedServer.Session.AccessToken = resp.GetAccessToken()
	persistedServer.Session.RefreshToken = resp.GetRefreshToken()

	if err := writeConfig(); err != nil {
		return err
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Logged in to %s as %s\n", targetServer.Name, username)
	return err
}

func resolveLoginServer(args []string, opts loginOptions) (*client.Server, bool, error) {
	configureServer := len(args) > 0 || opts.address != "" || opts.caFile != "" || opts.certFile != "" || opts.keyFile != "" || opts.insecure
	if !configureServer {
		currentSrv, err := cfg.CurrentServer()
		if err != nil {
			return nil, false, err
		}

		copied := *currentSrv
		if currentSrv.TLSConfig != nil {
			tlsCopy := *currentSrv.TLSConfig
			copied.TLSConfig = &tlsCopy
		}
		if currentSrv.Session != nil {
			sessionCopy := *currentSrv.Session
			copied.Session = &sessionCopy
		}
		return &copied, false, nil
	}

	if len(args) == 0 {
		return nil, false, fmt.Errorf("server name is required when configuring a server during login")
	}
	if opts.address == "" {
		return nil, false, fmt.Errorf("--address is required when configuring a server during login")
	}

	tlsEnabled := opts.caFile != "" || opts.certFile != "" || opts.keyFile != "" || opts.insecure
	targetServer, err := buildServerConfig(args[0], opts.address, opts.caFile, opts.certFile, opts.keyFile, opts.insecure, tlsEnabled)
	if err != nil {
		return nil, false, err
	}

	return targetServer, opts.current, nil
}

var loginClientFactory = func(server *client.Server) (loginClient, error) {
	return newLoginClient(server)
}

func newLoginClient(server *client.Server) (*client.ClientSet, error) {
	if server == nil {
		return nil, fmt.Errorf("server is nil")
	}

	if server.TLSConfig == nil {
		return client.New(server.Address)
	}

	cfgForServer := &client.Config{
		Current: server.Name,
		Servers: []*client.Server{server},
	}

	return client.New(server.Address, client.WithTLSConfigFromCfg(cfgForServer))
}

func writeConfig() error {
	configPath := viper.GetViper().ConfigFileUsed()
	if configPath == "" {
		return fmt.Errorf("config file path is not set")
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("error creating config dir: %v", err)
	}

	b, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("error marshal: %v", err)
	}

	if err := os.WriteFile(configPath, b, 0o666); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

func readLine(out io.Writer, in io.Reader, prompt string) (string, error) {
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return "", err
	}

	reader := bufio.NewReader(in)
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("value cannot be empty")
	}

	return value, nil
}

func readPassword(out io.Writer, prompt string) (string, error) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		if _, err := fmt.Fprint(out, prompt); err != nil {
			return "", err
		}

		password, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", err
		}
		if _, err := fmt.Fprintln(out); err != nil {
			return "", err
		}

		trimmed := strings.TrimSpace(string(password))
		if trimmed == "" {
			return "", fmt.Errorf("password cannot be empty")
		}

		return trimmed, nil
	}

	return readLine(out, os.Stdin, prompt)
}
