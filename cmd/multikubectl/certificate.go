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

	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/cmdutil"
)

func newCertificateCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "certificate",
		Aliases: []string{"cert", "certs", "certififcates"},
		Short:   "Manage certificates",
		Long:    `Manage certificates`,
	}

	cmd.AddCommand(newCertificateListCmd(cfg))
	cmd.AddCommand(newCertificateCreateCmd(cfg))

	return cmd
}

func newCertificateListCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List certificates registered with the server",
		Long: `Retrieve and display all certificates currently registered with
the server.`,
		Example: `  # List all certificates using the default configuration
  multikubectl certificate list

  # List certificates using a custom configuration file
  multikubectl certificate list --config /etc/multikube/config.yaml`,
		Args: cobra.NoArgs,
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runCertificateListCmd(cmd, cfg)
		}),
	}

	return cmd
}

func newCertificateCreateCmd(cfg *client.Config) *cobra.Command {
	var (
		certificate     string
		certificateData string
		key             string
		keyData         string
		labels          []string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new certificate",
		Long:  `Create a new certificate and register it with the server.`,
		Example: `  # Create a certificate with inline PEM certificate and key
  multikubectl certificate create my-cert \
    --certificate "$(cat tls.crt)" \
    --key "$(cat tls.key)"

  # Create a certificate with file paths
  multikubectl certificate create my-cert \
    --certificate-data /etc/ssl/tls.crt \
    --key-data /etc/ssl/tls.key

  # Create a certificate with labels
  multikubectl certificate create my-cert \
    --certificate-data /etc/ssl/tls.crt \
    --key-data /etc/ssl/tls.key \
    --label env=production --label team=platform`,
		Args: cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runCertificateCreateCmd(cmd, args, cfg, certificate, certificateData, key, keyData, labels)
		}),
	}

	cmd.Flags().StringVar(&certificate, "certificate", "", "PEM-encoded certificate (inline)")
	cmd.Flags().StringVar(&certificateData, "certificate-data", "", "Path to the PEM-encoded certificate file")
	cmd.Flags().StringVar(&key, "key", "", "PEM-encoded private key (inline)")
	cmd.Flags().StringVar(&keyData, "key-data", "", "Path to the PEM-encoded private key file")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Labels to attach in key=value format (can be specified multiple times)")

	return cmd
}

// runCertificateCreateCmd creates a new certificate via the multikube API server.
func runCertificateCreateCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
	certificate, certificateData, key, keyData string,
	labelStrs []string,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.certificate.create")
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

	cert := &certificatev1.Certificate{
		Meta: &metav1.Meta{
			Name:   name,
			Labels: cmdutil.ConvertKVStringsToMap(labelStrs),
		},
		Config: &certificatev1.CertificateConfig{
			Name:            name,
			Certificate:     certificate,
			CertificateData: certificateData,
			Key:             key,
			KeyData:         keyData,
		},
	}

	if err := c.CertificateV1().Create(ctx, cert); err != nil {
		logrus.Fatalf("error creating certificate: %v", err)
	}

	fmt.Printf("certificate %q created\n", name)

	return nil
}

// runCertificateListCmd lists all certificates registered with the multikube
// API server and prints them as a formatted table to stdout.
func runCertificateListCmd(cmd *cobra.Command, cfg *client.Config) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.certificate.list")
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

	certs, err := c.CertificateV1().List(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	_, _ = fmt.Fprintf(wr, "%s\t%s\t%s\n", "NAME", "GENERATION", "AGE")
	for _, cert := range certs {
		_, _ = fmt.Fprintf(wr, "%s\t%d\t%s\n",
			cert.GetMeta().GetName(),
			cert.GetMeta().GetGeneration(),
			cmdutil.FormatDuration(time.Since(cert.GetMeta().GetCreated().AsTime())),
		)
	}

	_ = wr.Flush()

	return nil
}
