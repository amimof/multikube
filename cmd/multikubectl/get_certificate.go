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

func newGetCertificateCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "certificate [NAME]",
		Aliases: []string{"cert", "certs", "certificates"},
		Short:   "Get certificates",
		Long:    `Retrieve and display certificates`,
		Args:    cobra.MaximumNArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runGetCertificateCmd(cmd, cfg, args[0])
			}
			return runListCertificatesCmd(cmd, cfg)
		}),
	}
	return cmd
}

// runCertificateCmd lists all certificates registered with the multikube API server
// and prints them as a formatted table to stdout.
func runGetCertificateCmd(cmd *cobra.Command, cfg *client.Config, name string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.certificate.list")
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

	lease, err := c.CertificateV1().Get(ctx, name)
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

// runListCertificatesCmd lists all certificates registered with the multikube API server
// and prints them as a formatted table to stdout.
func runListCertificatesCmd(cmd *cobra.Command, cfg *client.Config) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.certificate.list")
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

	tasks, err := c.CertificateV1().List(ctx)
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
