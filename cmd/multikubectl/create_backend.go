package main

import (
	"context"
	"fmt"
	"time"

	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/cmdutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/types/known/durationpb"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
)

func newCreateBackendCmd(cfg *client.Config) *cobra.Command {
	var (
		server          []string
		caRef           string
		authRef         string
		insecureSkipTLS bool
		cacheTTL        time.Duration
		labels          []string
	)

	cmd := &cobra.Command{
		Use:   "backend [NAME]",
		Short: "Create a new backend",
		Long:  `Create a new backend and register it with the server.`,
		Args:  cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runCreateBackendCmd(cmd, args, cfg, server, caRef, authRef, insecureSkipTLS, cacheTTL, labels)
		}),
	}

	cmd.Flags().StringArrayVar(&server, "server", []string{}, "Address of the Kubernetes API server for this backend (required)")
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

// runCreateBackendCmd creates a new backend via the multikube API server.
func runCreateBackendCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
	server []string, caRef, authRef string,
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
			Servers:               server,
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
