package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/types/known/durationpb"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/cmdutil"
)

func newBackendCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backend",
		Short: "Manage backends registered with the multikube API server",
		Long: `Manage backends registered with the multikube API server.

Use subcommands to list or create backends.`,
	}

	cmd.AddCommand(newBackendListCmd(cfg))
	cmd.AddCommand(newBackendCreateCmd(cfg))

	return cmd
}

func newBackendListCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backends registered with the multikube API server",
		Long: `Retrieve and display all backends currently registered with the multikube
API server.

The command connects to the API server whose address is resolved from the
current server entry in the multikube configuration file and prints a
summary table with each backend's name, generation, and age.`,
		Example: `  # List all backends using the default configuration
  multikubectl backend list

  # List backends from a non-default API server address
  multikubectl backend list --server 10.0.0.1:5700

  # List backends using a custom configuration file
  multikubectl backend list --config /etc/multikube/config.yaml`,
		Args: cobra.NoArgs,
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runBackendCmd(cmd, cfg)
		}),
	}

	return cmd
}

func newBackendCreateCmd(cfg *client.Config) *cobra.Command {
	var (
		server          string
		caRef           string
		authRef         string
		insecureSkipTLS bool
		cacheTTL        time.Duration
		labels          []string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new backend",
		Long: `Create a new backend and register it with the multikube API server.

The NAME argument is required and sets the backend's name. The --server flag
is required and specifies the Kubernetes API server address for this backend.
All other fields are optional and have sensible defaults or can be left empty.`,
		Example: `  # Create a backend with a required name and server
  multikubectl backend create my-cluster --server https://k8s.example.com:6443

  # Create a backend with TLS references
  multikubectl backend create my-cluster --server https://k8s.example.com:6443 \
    --ca-ref my-ca-secret --auth-ref my-auth-secret

  # Create a backend skipping TLS verification
  multikubectl backend create my-cluster --server https://k8s.example.com:6443 \
    --insecure-skip-tls-verify

  # Create a backend with a cache TTL of 5 minutes
  multikubectl backend create my-cluster --server https://k8s.example.com:6443 \
    --cache-ttl 5m`,
		Args: cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runBackendCreateCmd(cmd, args, cfg, server, caRef, authRef, insecureSkipTLS, cacheTTL, labels)
		}),
	}

	cmd.Flags().StringVar(&server, "server", "", "Address of the Kubernetes API server for this backend (required)")
	cmd.Flags().StringVar(&caRef, "ca-ref", "", "Reference to the CA certificate secret")
	cmd.Flags().StringVar(&authRef, "auth-ref", "", "Reference to the authentication secret")
	cmd.Flags().BoolVar(&insecureSkipTLS, "insecure-skip-tls-verify", false, "Skip TLS certificate verification for the backend server")
	cmd.Flags().DurationVar(&cacheTTL, "cache-ttl", 0, "Cache time-to-live duration (e.g. 30s, 5m, 1h). Zero means no caching.")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Labels to attach in key=value format (can be specified multiple times)")

	if err := cmd.MarkFlagRequired("server"); err != nil {
		logrus.Fatalf("error marking flag as required: %v", err)
	}

	return cmd
}

// runBackendCreateCmd creates a new backend via the multikube API server.
func runBackendCreateCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
	server, caRef, authRef string,
	insecureSkipTLS bool,
	cacheTTL time.Duration,
	labelStrs []string,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.backend.create")
	defer span.End()

	name := args[0]

	// Setup client
	currentSrv, err := cfg.CurrentServer()
	if err != nil {
		logrus.Fatal(err)
	}
	c, err := client.New(currentSrv.Address, client.WithTLSConfigFromCfg(cfg))
	if err != nil {
		logrus.Fatalf("error setting up client: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			logrus.Errorf("error closing client connection: %v", err)
		}
	}()

	// Build the backend object
	backend := &backendv1.Backend{
		Meta: &metav1.Meta{
			Name:   name,
			Labels: cmdutil.ConvertKVStringsToMap(labelStrs),
		},
		Config: &backendv1.BackendConfig{
			Name:                  name,
			Server:                server,
			CaRef:                 caRef,
			AuthRef:               authRef,
			InsecureSkipTlsVerify: insecureSkipTLS,
		},
	}

	if cacheTTL > 0 {
		backend.Config.CacheTtl = durationpb.New(cacheTTL)
	}

	if err := c.BackendV1().Create(ctx, backend); err != nil {
		logrus.Fatalf("error creating backend: %v", err)
	}

	fmt.Printf("backend %q created\n", name)

	return nil
}

// runBackendCmd lists all backends registered with the multikube API server
// and prints them as a formatted table to stdout.
func runBackendCmd(cmd *cobra.Command, cfg *client.Config) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.backend.list")
	defer span.End()

	// Setup client
	currentSrv, err := cfg.CurrentServer()
	if err != nil {
		logrus.Fatal(err)
	}
	c, err := client.New(currentSrv.Address, client.WithTLSConfigFromCfg(cfg))
	if err != nil {
		logrus.Fatalf("error setting up client: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			logrus.Errorf("error closing client connection: %v", err)
		}
	}()
	// Setup writer
	wr := tabwriter.NewWriter(os.Stdout, 8, 8, 8, '\t', tabwriter.AlignRight)

	tasks, err := c.BackendV1().List(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	_, _ = fmt.Fprintf(wr, "%s\t%s\t%s\n", "NAME", "GENERATION", "AGE")
	for _, c := range tasks {
		_, _ = fmt.Fprintf(wr, "%s\t%d\t%s\n",
			c.GetMeta().GetName(),
			c.GetMeta().GetGeneration(),
			cmdutil.FormatDuration(time.Since(c.GetMeta().GetCreated().AsTime())),
		)
	}

	_ = wr.Flush()

	return nil
}
