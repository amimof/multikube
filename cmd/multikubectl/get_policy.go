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

func newGetPolicyCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "policy [NAME]",
		Short:   "Get policies",
		Long:    `Retrieve and display policies`,
		Aliases: []string{"policies"},
		Args:    cobra.MaximumNArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runGetPolicyCmd(cmd, cfg, args[0])
			}
			return runListPoliciesCmd(cmd, cfg)
		}),
	}
	return cmd
}

func runGetPolicyCmd(cmd *cobra.Command, cfg *client.Config, name string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.policy.get")
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

	policy, err := c.PolicyV1().Get(ctx, name)
	if err != nil {
		logrus.Fatal(err)
	}

	codec, err := cmdutil.CodecFor(outputFormat)
	if err != nil {
		logrus.Fatalf("error creating serializer: %v", err)
	}

	b, err := codec.Serialize(policy)
	if err != nil {
		logrus.Fatalf("error serializing: %v", err)
	}

	fmt.Printf("%s\n", string(b))

	return nil
}

func runListPoliciesCmd(cmd *cobra.Command, cfg *client.Config) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.policy.list")
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

	policies, err := c.PolicyV1().List(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	_, _ = fmt.Fprintf(wr, "%s\t%s\t%s\n", "NAME", "GENERATION", "AGE")
	for _, policy := range policies {
		_, _ = fmt.Fprintf(wr, "%s\t%d\t%s\n",
			policy.GetMeta().GetName(),
			policy.GetMeta().GetGeneration(),
			cmdutil.FormatDuration(time.Since(policy.GetMeta().GetCreated().AsTime())),
		)
	}

	_ = wr.Flush()

	return nil
}
