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

func newDeleteCertificateAuthorityCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ca [NAME]",
		Short: "Delete one or more cas",
		Long:  `Delete one or more cas`,
		Example: `
  # Delete a ca
  multikubectl delete ca prod-cert-v1

  # Delete many cas
  multikubectl delete ca prod-cert-v1 cloud-cert external-clients-cert`,
		Args: cobra.ExactArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runDeleteCertificateAuthorityCmd(cmd, args, cfg)
		}),
	}

	return cmd
}

// runDeleteCertificateAuthorityCmd deletes a new ca
func runDeleteCertificateAuthorityCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.ca.delete")
	defer span.End()

	name := args[0]

	if err := clientSet.CAV1().Delete(ctx, name); err != nil {
		logrus.Fatalf("error creating ca: %v", err)
	}

	fmt.Printf("ca %q deleted\n", name)

	return nil
}
