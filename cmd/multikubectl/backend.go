package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"

	"github.com/amimof/multikube/pkg/client"
)

func newBackendCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backend",
		Short: "List backends registered with the multikube API server",
		Long: `Retrieve and display all backends currently registered with the multikube
API server.

The command connects to the API server whose address is resolved from the
current server entry in the multikube configuration file and prints a
summary table with each backend's name, generation, phase, reason, node,
and age.`,
		Example: `  # List all backends using the default configuration
  multikubectl backend

  # List backends from a non-default API server address
  multikubectl backend --server 10.0.0.1:5700

  # List backends using a custom configuration file
  multikubectl backend --config /etc/multikube/config.yaml`,
		Args: cobra.NoArgs,
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runBackendCmd(cmd, cfg)
		}),
	}

	return cmd
}

// runBackendCmd lists all backends registered with the multikube API server
// and prints them as a formatted table to stdout.
func runBackendCmd(cmd *cobra.Command, cfg *client.Config) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.backend.list")
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
	// Setup writer
	wr := tabwriter.NewWriter(os.Stdout, 8, 8, 8, '\t', tabwriter.AlignRight)

	tasks, err := c.BackendV1().List(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	_, _ = fmt.Fprintf(wr, "%s\t%s\t%s\t%s\t%s\t%s\n", "NAME", "GENERATION", "PHASE", "REASON", "NODE", "AGE")
	for _, c := range tasks {
		_, _ = fmt.Fprintf(wr, "%s\t%d\t%s\t%s\t%s\t%s\n",
			c.GetMeta().GetName(),
			c.GetMeta().GetGeneration(),
			// c.GetStatus().GetReason().GetValue(),
			// c.GetStatus().GetNode().GetValue(),
			// cmdutil.FormatDuration(time.Since(c.GetMeta().GetCreated().AsTime())),
		)
	}

	_ = wr.Flush()

	return nil
}
