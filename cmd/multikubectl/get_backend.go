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

func newGetBackendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "backend [NAME]",
		Short:   "Get backends",
		Long:    `Retrieve and display backends`,
		Aliases: []string{"backends"},
		Args:    cobra.MaximumNArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
			defer cancel()

			if len(args) == 1 {
				return runGetBackendCmd(ctx, cmd, args[0])
			}
			return runListBackendsCmd(ctx, cmd)
		}),
	}
	return cmd
}

// runBackendCmd lists all backends registered with the multikube API server
// and prints them as a formatted table to stdout.
func runGetBackendCmd(ctx context.Context, cmd *cobra.Command, name string) error {
	lease, err := clientSet.BackendV1().Get(ctx, name)
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

// runListBackendsCmd lists all backends registered with the multikube API server
// and prints them as a formatted table to stdout.
func runListBackendsCmd(ctx context.Context, cmd *cobra.Command) error {
	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.backend.list")
	defer span.End()

	// Setup writer
	wr := tabwriter.NewWriter(os.Stdout, 8, 8, 8, '\t', tabwriter.AlignRight)

	tasks, err := clientSet.BackendV1().List(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	_, _ = fmt.Fprintf(wr, "%s\t%s\t%s\n", "NAME", "GENERATION", "AGE")
	for _, c := range tasks {
		_, _ = fmt.Fprintf(wr, "%s\t%d\t%s\n",
			c.GetMeta().GetName(),
			c.GetMeta().GetGeneration(),
			cmdutil.FormatDuration(time.Since(c.GetMeta().GetCreated().AsTime())),
		)
	}

	_ = wr.Flush()

	return nil
}
