package main

import (
	"context"
	"fmt"
	"time"

	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/cmdutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

func newCreateCertificateCmd(cfg *client.Config) *cobra.Command {
	var (
		certificate     string
		certificateData string
		key             string
		keyData         string
		labels          []string
	)

	cmd := &cobra.Command{
		Use:     "certificate [NAME]",
		Aliases: []string{"cert"},
		Short:   "Create a new certificate",
		Long:    `Create a new certificate and register it with the server.`,
		Example: `  # Create a certificate with inline PEM certificate and key
  multikubectl create certificate my-cert \
    --certificate "$(cat tls.crt)" \
    --key "$(cat tls.key)"

  # Create a certificate with file paths
  multikubectl create certificate my-cert \
    --certificate-data /etc/ssl/tls.crt \
    --key-data /etc/ssl/tls.key

  # Create a certificate with labels
  multikubectl create certificate my-cert \
    --certificate-data /etc/ssl/tls.crt \
    --key-data /etc/ssl/tls.key \
    --label env=production --label team=platform`,
		Args: cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runCreateCertificateCmd(cmd, args, cfg, certificate, certificateData, key, keyData, labels)
		}),
	}

	cmd.Flags().StringVar(&certificate, "certificate", "", "PEM-encoded certificate (inline)")
	cmd.Flags().StringVar(&certificateData, "certificate-data", "", "Path to the PEM-encoded certificate file")
	cmd.Flags().StringVar(&key, "key", "", "PEM-encoded private key (inline)")
	cmd.Flags().StringVar(&keyData, "key-data", "", "Path to the PEM-encoded private key file")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Labels to attach in key=value format (can be specified multiple times)")

	return cmd
}

// runCreateCertificateCmd creates a new certificate
func runCreateCertificateCmd(
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
