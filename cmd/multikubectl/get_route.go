package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/cmdutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

func newGetRouteCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "route [NAME]",
		Short:   "Get routes",
		Long:    `Retrieve and display routes`,
		Aliases: []string{"routes"},
		Args:    cobra.MaximumNArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runGetRouteCmd(cmd, cfg, args[0])
			}
			return runListRoutesCmd(cmd, cfg)
		}),
	}
	return cmd
}

// runRouteCmd lists all routes registered with the server
func runGetRouteCmd(cmd *cobra.Command, cfg *client.Config, name string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.route.list")
	defer span.End()

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

	lease, err := c.RouteV1().Get(ctx, name)
	if err != nil {
		logrus.Fatal(err)
	}

	codec, err := cmdutil.CodecFor(outputFormat)
	if err != nil {
		logrus.Fatalf("error creating serializer: %v", err)
	}

	b, err := codec.Serialize(lease)
	if err != nil {
		logrus.Fatalf("error serializing: %v", err)
	}

	fmt.Printf("%s\n", string(b))

	return nil
}

// runRouteListCmd lists all routes registered with the server
func runListRoutesCmd(cmd *cobra.Command, cfg *client.Config) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.route.list")
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

	routes, err := c.RouteV1().List(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	_, _ = fmt.Fprintf(wr, "%s\t%s\t%s\t%s\n", "NAME", "PHASE", "GENERATION", "AGE")
	for _, route := range routes {
		_, _ = fmt.Fprintf(wr, "%s\t%s\t%d\t%s\n",
			route.GetMeta().GetName(),
			route.GetStatus().GetPhase().GetValue(),
			route.GetMeta().GetGeneration(),
			cmdutil.FormatDuration(time.Since(route.GetMeta().GetCreated().AsTime())),
		)
	}

	_ = wr.Flush()

	return nil
}
