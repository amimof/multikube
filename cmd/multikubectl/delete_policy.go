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

func newDeletePolicyCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy [NAME]",
		Short: "Delete one or more policys",
		Long:  `Delete one or more policys`,
		Example: `
  # Delete a policy
  multikubectl delete policy prod-cert-v1

  # Delete many policys
  multikubectl delete policy prod-cert-v1 cloud-cert external-clients-cert`,
		Args: cobra.ExactArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runDeletePolicyCmd(cmd, args, cfg)
		}),
	}

	return cmd
}

// runDeletePolicyCmd deletes a new policy
func runDeletePolicyCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.policy.delete")
	defer span.End()

	name := args[0]

	if err := clientSet.PolicyV1().Delete(ctx, name); err != nil {
		logrus.Fatalf("error creating policy: %v", err)
	}

	fmt.Printf("policy %q deleted\n", name)

	return nil
}
