package main

import (
	"context"
	"os"
	"time"

	policyv1 "github.com/amimof/multikube/api/policy/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

func newEditPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy [NAME]",
		Short: "Edit a policy",
		Long:  `Edit a policy`,
		Example: `
  # Edit a policy
  multikubectl edit policy prod-policy-v1`,
		Args: cobra.ExactArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runEditPolicyCmd(cmd, args)
		}),
	}

	return cmd
}

func runEditPolicyCmd(cmd *cobra.Command, args []string) error {
	baseCtx := cmd.Context()

	tracer := otel.Tracer("multikubectl")
	baseCtx, span := tracer.Start(baseCtx, "multikubectl.edit.policy")
	defer span.End()

	nname := args[0]

	getCtx, policyncel := context.WithTimeout(baseCtx, time.Second*30)
	defer policyncel()

	policy, err := clientSet.PolicyV1().Get(getCtx, nname)
	if err != nil {
		return err
	}

	var updatedPolicy policyv1.Policy
	err = runEditor(policy, &updatedPolicy)
	if err != nil {
		return err
	}

	// Exit early if no changes where made
	if proto.Equal(policy, &updatedPolicy) {
		logrus.Info("no changes detected")
		os.Exit(0)
	}

	// Send update to server
	updateCtx, policyncel := context.WithTimeout(baseCtx, time.Second*30)
	defer policyncel()
	err = clientSet.PolicyV1().Update(updateCtx, nname, &updatedPolicy)
	if err != nil {
		return err
	}

	logrus.Infof("policy %s was updated", nname)

	return nil
}
