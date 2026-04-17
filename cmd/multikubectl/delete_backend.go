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

func newDeleteBackendCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backend [NAME]",
		Short: "Delete one or more backends",
		Long:  `Delete one or more backends`,
		Example: `
  # Delete a backend
  multikubectl delete backend prod-cert-v1

  # Delete many backends
  multikubectl delete backend prod-cert-v1 cloud-cert external-clients-cert`,
		Args: cobra.ExactArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runDeleteBackendCmd(cmd, args, cfg)
		}),
	}

	return cmd
}

// runDeleteBackendCmd deletes a new backend
func runDeleteBackendCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.backend.delete")
	defer span.End()

	name := args[0]

	if err := clientSet.BackendV1().Delete(ctx, name); err != nil {
		logrus.Fatalf("error creating backend: %v", err)
	}

	fmt.Printf("backend %q deleted\n", name)

	return nil
}
