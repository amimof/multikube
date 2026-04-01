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

func newDeleteRouteCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route [NAME]",
		Short: "Delete one or more routes",
		Long:  `Delete one or more routes`,
		Example: `
  # Delete a route
  multikubectl delete route prod-cert-v1

  # Delete many routes
  multikubectl delete route prod-cert-v1 cloud-cert external-clients-cert`,
		Args: cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runDeleteRouteCmd(cmd, args, cfg)
		}),
	}

	return cmd
}

// runDeleteRouteCmd deletes a new route
func runDeleteRouteCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.route.delete")
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

	if err := c.RouteV1().Delete(ctx, name); err != nil {
		logrus.Fatalf("error creating route: %v", err)
	}

	fmt.Printf("route %q deleted\n", name)

	return nil
}
