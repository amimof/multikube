package main

import (
	"context"
	"fmt"
	"time"

	"github.com/amimof/multikube/pkg/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

func newDeleteCertificateCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certificate [NAME]",
		Short: "Delete one or more certificates",
		Long:  `Delete one or more certificates`,
		Example: `
  # Delete a certificate
  multikubectl delete certificate prod-cert-v1

  # Delete many certificates
  multikubectl delete certificate prod-cert-v1 cloud-cert external-clients-cert`,
		Args: cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runDeleteCertificateCmd(cmd, args, cfg)
		}),
	}

	return cmd
}

// runDeleteCertificateCmd deletes a new certificate
func runDeleteCertificateCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.certificate.delete")
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

	if err := c.CertificateV1().Delete(ctx, name); err != nil {
		logrus.Fatalf("error creating certificate: %v", err)
	}

	fmt.Printf("certificate %q deleted\n", name)

	return nil
}
