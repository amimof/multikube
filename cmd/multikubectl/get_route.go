package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/amimof/multikube/pkg/cmdutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

func newGetRouteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "route [NAME]",
		Short:   "Get routes",
		Long:    `Retrieve and display routes`,
		Aliases: []string{"routes"},
		Args:    cobra.MaximumNArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
			defer cancel()
			if len(args) == 1 {
				return runGetRouteCmd(ctx, cmd, args[0])
			}
			return runListRoutesCmd(ctx, cmd)
		}),
	}
	return cmd
}

// runRouteCmd lists all routes registered with the server
func runGetRouteCmd(ctx context.Context, cmd *cobra.Command, name string) error {
	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.route.list")
	defer span.End()

	lease, err := clientSet.RouteV1().Get(ctx, name)
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
func runListRoutesCmd(ctx context.Context, cmd *cobra.Command) error {
	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.route.list")
	defer span.End()

	wr := tabwriter.NewWriter(os.Stdout, 8, 8, 8, '\t', tabwriter.AlignRight)

	routes, err := clientSet.RouteV1().List(ctx)
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
