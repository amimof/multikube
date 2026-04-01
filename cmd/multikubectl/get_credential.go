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

func newGetCredentialCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "credential [NAME]",
		Short:   "Get credentials",
		Long:    `Retrieve and display credentials`,
		Aliases: []string{"credentials"},
		Args:    cobra.MaximumNArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runGetCredentialCmd(cmd, cfg, args[0])
			}
			return runListCredentialsCmd(cmd, cfg)
		}),
	}

	return cmd
}

func runGetCredentialCmd(cmd *cobra.Command, cfg *client.Config, name string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.credential.get")
	defer span.End()

	currentSrv, err := cfg.CurrentServer()
	if err != nil {
		logrus.Fatal(err)
	}
	clientset, err := client.New(currentSrv.Address, client.WithTLSConfigFromCfg(cfg))
	if err != nil {
		logrus.Fatalf("error setting up client: %v", err)
	}
	defer func() {
		if err := clientset.Close(); err != nil {
			logrus.Errorf("error closing client connection: %v", err)
		}
	}()

	credential, err := clientset.CredentialV1().Get(ctx, name)
	if err != nil {
		logrus.Fatal(err)
	}

	codec, err := cmdutil.CodecFor(outputFormat)
	if err != nil {
		logrus.Fatalf("error creating serializer: %v", err)
	}

	b, err := codec.Serialize(credential)
	if err != nil {
		logrus.Fatalf("error serializing: %v", err)
	}

	fmt.Printf("%s\n", string(b))

	return nil
}

func runListCredentialsCmd(cmd *cobra.Command, cfg *client.Config) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.credential.list")
	defer span.End()

	currentSrv, err := cfg.CurrentServer()
	if err != nil {
		logrus.Fatal(err)
	}
	clientset, err := client.New(currentSrv.Address, client.WithTLSConfigFromCfg(cfg))
	if err != nil {
		logrus.Fatalf("error setting up client: %v", err)
	}
	defer func() {
		if err := clientset.Close(); err != nil {
			logrus.Errorf("error closing client connection: %v", err)
		}
	}()

	credentials, err := clientset.CredentialV1().List(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	wr := tabwriter.NewWriter(os.Stdout, 8, 8, 8, '\t', tabwriter.AlignRight)
	_, _ = fmt.Fprintf(wr, "%s\t%s\t%s\n", "NAME", "GENERATION", "AGE")
	for _, credential := range credentials {
		_, _ = fmt.Fprintf(wr, "%s\t%d\t%s\n",
			credential.GetMeta().GetName(),
			credential.GetMeta().GetGeneration(),
			cmdutil.FormatDuration(time.Since(credential.GetMeta().GetCreated().AsTime())),
		)
	}

	_ = wr.Flush()

	return nil
}
