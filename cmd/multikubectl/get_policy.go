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

func newGetPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "policy [NAME]",
		Short:   "Get policies",
		Long:    `Retrieve and display policies`,
		Aliases: []string{"policies"},
		Args:    cobra.MaximumNArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
			defer cancel()
			if len(args) == 1 {
				return runGetPolicyCmd(ctx, cmd, args[0])
			}
			return runListPoliciesCmd(ctx, cmd)
		}),
	}
	return cmd
}

func runGetPolicyCmd(ctx context.Context, cmd *cobra.Command, name string) error {
	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.policy.get")
	defer span.End()

	policy, err := clientSet.PolicyV1().Get(ctx, name)
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

func runListPoliciesCmd(ctx context.Context, cmd *cobra.Command) error {
	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.policy.list")
	defer span.End()

	wr := tabwriter.NewWriter(os.Stdout, 8, 8, 8, '\t', tabwriter.AlignRight)

	policies, err := clientSet.PolicyV1().List(ctx)
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
