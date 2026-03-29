package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"

	metav1 "github.com/amimof/multikube/api/meta/v1"
	policyv1 "github.com/amimof/multikube/api/policy/v1"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/cmdutil"
)

func newCreatePolicyCmd(cfg *client.Config) *cobra.Command {
	var labels []string

	cmd := &cobra.Command{
		Use:   "policy [NAME]",
		Short: "Create a new policy",
		Long: `Create a new policy and register it with the server.

The NAME argument is required and sets the policy's name.`,
		Example: `  # Create an empty policy
  multikubectl create policy my-policy

  # Create a policy with labels
  multikubectl create policy my-policy --label env=production`,
		Args: cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runCreatePolicyCmd(cmd, args, cfg, labels)
		}),
	}

	cmd.Flags().StringArrayVar(&labels, "label", nil, "Labels to attach in key=value format (can be specified multiple times)")

	return cmd
}

func runCreatePolicyCmd(cmd *cobra.Command, args []string, cfg *client.Config, labelStrs []string) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.policy.create")
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

	policy := &policyv1.Policy{
		Meta: &metav1.Meta{
			Name:   name,
			Labels: cmdutil.ConvertKVStringsToMap(labelStrs),
		},
		Config: &policyv1.PolicyConfig{
			Name: name,
		},
	}

	if err := c.PolicyV1().Create(ctx, policy); err != nil {
		logrus.Fatalf("error creating policy: %v", err)
	}

	fmt.Printf("policy %q created\n", name)

	return nil
}
