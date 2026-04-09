package main

import (
	"context"
	"os"
	"time"

	credentialv1 "github.com/amimof/multikube/api/credential/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

func newEditCredentialCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "credential [NAME]",
		Short: "Edit a credential",
		Long:  `Edit a credential`,
		Example: `
  # Edit a credential
  multikubectl edit credential prod-credential-v1`,
		Args: cobra.ExactArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runEditCredentialCmd(cmd, args)
		}),
	}

	return cmd
}

func runEditCredentialCmd(cmd *cobra.Command, args []string) error {
	baseCtx := cmd.Context()

	tracer := otel.Tracer("multikubectl")
	baseCtx, span := tracer.Start(baseCtx, "multikubectl.edit.credential")
	defer span.End()

	nname := args[0]

	getCtx, credentialncel := context.WithTimeout(baseCtx, time.Second*30)
	defer credentialncel()

	credential, err := clientSet.CredentialV1().Get(getCtx, nname)
	if err != nil {
		return err
	}

	var updatedCredential credentialv1.Credential
	err = runEditor(credential, &updatedCredential)
	if err != nil {
		return err
	}

	// Exit early if no changes where made
	if proto.Equal(credential, &updatedCredential) {
		logrus.Info("no changes detected")
		os.Exit(0)
	}

	// Send update to server
	updateCtx, credentialncel := context.WithTimeout(baseCtx, time.Second*30)
	defer credentialncel()
	err = clientSet.CredentialV1().Update(updateCtx, nname, &updatedCredential)
	if err != nil {
		return err
	}

	logrus.Infof("credential %s was updated", nname)

	return nil
}
