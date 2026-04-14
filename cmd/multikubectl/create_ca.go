package main

import (
	"context"
	"fmt"
	"time"

	cav1 "github.com/amimof/multikube/api/ca/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/cmdutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

func newCreateCACmd(cfg *client.Config) *cobra.Command {
	var (
		certificate     string
		certificateData string
		labels          []string
	)

	cmd := &cobra.Command{
		Use:   "ca [NAME]",
		Short: "Create a new certificate authority",
		Long:  `Create a new certificate authority and register it with the server.`,
		Example: `  # Create a CA with an inline PEM certificate
  multikubectl create ca my-ca --certificate "$(cat ca.crt)"

  # Create a CA with a certificate file path
  multikubectl create ca my-ca --certificate-data /etc/ssl/ca.crt

  # Create a CA with labels
  multikubectl create ca my-ca --certificate-data /etc/ssl/ca.crt \
    --label env=production --label team=platform`,
		Args: cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runCreateCACmd(cmd, args, cfg, certificate, certificateData, labels)
		}),
	}

	cmd.Flags().StringVar(&certificate, "certificate", "", "PEM-encoded CA certificate (inline)")
	cmd.Flags().StringVar(&certificateData, "certificate-data", "", "Path to the PEM-encoded CA certificate file")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Labels to attach in key=value format (can be specified multiple times)")

	return cmd
}

// runCreateCACmd creates a new certificate authority
func runCreateCACmd(
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
