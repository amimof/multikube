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

	cav1 "github.com/amimof/multikube/api/ca/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/cmdutil"
)

func newCACmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ca",
		Aliases: []string{"cas"},
		Short:   "Manage certificate authorities",
		Long: `Manage certificate authorities.

Use subcommands to list or create certificate authorities.`,
	}

	cmd.AddCommand(newCAListCmd(cfg))
	cmd.AddCommand(newCACreateCmd(cfg))

	return cmd
}

func newCAListCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List certificate authorities registered with the server",
		Long: `Retrieve and display all certificate authorities currently registered with
the server.`,
		Example: `  # List all certificate authorities using the default configuration
  multikubectl ca list

  # List certificate authorities from a non-default API server address
  multikubectl ca list --server 10.0.0.1:5700

  # List certificate authorities using a custom configuration file
  multikubectl ca list --config /etc/multikube/config.yaml`,
		Args: cobra.NoArgs,
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runCAListCmd(cmd, cfg)
		}),
	}

	return cmd
}

func newCACreateCmd(cfg *client.Config) *cobra.Command {
	var (
		certificate     string
		certificateData string
		labels          []string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new certificate authority",
		Long:  `Create a new certificate authority and register it with the server.`,
		Example: `  # Create a CA with an inline PEM certificate
  multikubectl ca create my-ca --certificate "$(cat ca.crt)"

  # Create a CA with a certificate file path
  multikubectl ca create my-ca --certificate-data /etc/ssl/ca.crt

  # Create a CA with labels
  multikubectl ca create my-ca --certificate-data /etc/ssl/ca.crt \
    --label env=production --label team=platform`,
		Args: cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runCACreateCmd(cmd, args, cfg, certificate, certificateData, labels)
		}),
	}

	cmd.Flags().StringVar(&certificate, "certificate", "", "PEM-encoded CA certificate (inline)")
	cmd.Flags().StringVar(&certificateData, "certificate-data", "", "Path to the PEM-encoded CA certificate file")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Labels to attach in key=value format (can be specified multiple times)")

	return cmd
}

// runCACreateCmd creates a new certificate authority
func runCACreateCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
	certificate, certificateData string,
	labelStrs []string,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.ca.create")
	defer span.End()

	name := args[0]

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

	ca := &cav1.CertificateAuthority{
		Meta: &metav1.Meta{
			Name:   name,
			Labels: cmdutil.ConvertKVStringsToMap(labelStrs),
		},
		Config: &cav1.CertificateAuthorityConfig{
			Name:            name,
			Certificate:     certificate,
			CertificateData: certificateData,
		},
	}

	if err := c.CAV1().Create(ctx, ca); err != nil {
		logrus.Fatalf("error creating certificate authority: %v", err)
	}

	fmt.Printf("certificateauthority %q created\n", name)

	return nil
}

// runCAListCmd lists all certificate authorities registered with the multikube
// API server and prints them as a formatted table to stdout.
func runCAListCmd(cmd *cobra.Command, cfg *client.Config) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.ca.list")
	defer span.End()

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

	wr := tabwriter.NewWriter(os.Stdout, 8, 8, 8, '\t', tabwriter.AlignRight)

	cas, err := c.CAV1().List(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	_, _ = fmt.Fprintf(wr, "%s\t%s\t%s\n", "NAME", "GENERATION", "AGE")
	for _, ca := range cas {
		_, _ = fmt.Fprintf(wr, "%s\t%d\t%s\n",
			ca.GetMeta().GetName(),
			ca.GetMeta().GetGeneration(),
			cmdutil.FormatDuration(time.Since(ca.GetMeta().GetCreated().AsTime())),
		)
	}

	_ = wr.Flush()

	return nil
}
