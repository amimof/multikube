package main

import (
	"context"
	"fmt"
	"os"
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
		certificate string
		key         string
		labels      []string
	)

	cmd := &cobra.Command{
		Use:     "certificate [NAME]",
		Aliases: []string{"cert"},
		Short:   "Create a new certificate",
		Long:    `Create a new certificate and register it with the server.`,
		Example: `  # Create a certificate with inline PEM certificate and key
  multikubectl create certificate my-cert \
    --certificate /etc/ssl/tls.crt \
    --key /etc/ssl/tls.key 

  # Create a certificate with labels
  multikubectl create certificate my-cert \
    --certificate /etc/ssl/tls.crt \
    --key /etc/ssl/tls.key \
    --label env=production --label team=platform`,
		Args: cobra.ExactArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runCreateCertificateCmd(cmd, args, cfg, certificate, key, labels)
		}),
	}

	cmd.Flags().StringVar(&certificate, "certificate", "", "Path to PEM-encoded certificate")
	cmd.Flags().StringVar(&key, "key", "", "Path to PEM-encoded private key")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Labels to attach in key=value format (can be specified multiple times)")

	return cmd
}

func readFileFromPath(certPath string) ([]byte, error) {
	b, err := os.ReadFile(certPath)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// runCreateCertificateCmd creates a new certificate
func runCreateCertificateCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
	certificate, key string,
	labelStrs []string,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.certificate.create")
	defer span.End()

	name := args[0]

	certData, err := readFileFromPath(certificate)
	if err != nil {
		return err
	}
	keyData, err := readFileFromPath(key)
	if err != nil {
		return err
	}

	cert := &certificatev1.Certificate{
		Meta: &metav1.Meta{
			Name:   name,
			Labels: cmdutil.ConvertKVStringsToMap(labelStrs),
		},
		Config: &certificatev1.CertificateConfig{
			CertificateData: string(certData),
			KeyData:         string(keyData),
		},
	}

	if err := clientSet.CertificateV1().Create(ctx, cert); err != nil {
		logrus.Fatalf("error creating certificate: %v", err)
	}

	fmt.Printf("certificate %q created\n", name)

	return nil
}
