package main

import (
	"context"
	"os"
	"time"

	cav1 "github.com/amimof/multikube/api/ca/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

func newEditCACmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ca [NAME]",
		Short: "Edit a certificate authority",
		Long:  `Edit a certificate authority`,
		Example: `
  # Edit a ca
  multikubectl edit ca prod-ca-v1`,
		Args: cobra.ExactArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runEditCACmd(cmd, args)
		}),
	}

	return cmd
}

func runEditCACmd(cmd *cobra.Command, args []string) error {
	baseCtx := cmd.Context()

	tracer := otel.Tracer("multikubectl")
	baseCtx, span := tracer.Start(baseCtx, "multikubectl.edit.ca")
	defer span.End()

	nname := args[0]

	getCtx, cancel := context.WithTimeout(baseCtx, time.Second*30)
	defer cancel()

	ca, err := clientSet.CAV1().Get(getCtx, nname)
	if err != nil {
		return err
	}

	var updatedCA cav1.CertificateAuthority
	err = runEditor(ca, &updatedCA)
	if err != nil {
		return err
	}

	// Exit early if no changes where made
	if proto.Equal(ca, &updatedCA) {
		logrus.Info("no changes detected")
		os.Exit(0)
	}

	// Send update to server
	updateCtx, cancel := context.WithTimeout(baseCtx, time.Second*30)
	defer cancel()
	err = clientSet.CAV1().Update(updateCtx, nname, &updatedCA)
	if err != nil {
		return err
	}

	logrus.Infof("certificate authority %s was updated", nname)

	return nil
}
