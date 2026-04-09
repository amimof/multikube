package main

import (
	"context"
	"os"
	"time"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

func newEditBackendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backend [NAME]",
		Short: "Edit a backend",
		Long:  `Edit a backend`,
		Example: `
  # Edit a backend
  multikubectl edit backend prod-cert-v1`,
		Args: cobra.ExactArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runEditBackendCmd(cmd, args)
		}),
	}

	return cmd
}

func runEditBackendCmd(cmd *cobra.Command, args []string) error {
	baseCtx := cmd.Context()

	tracer := otel.Tracer("multikubectl")
	baseCtx, span := tracer.Start(baseCtx, "multikubectl.edit.ca")
	defer span.End()

	nname := args[0]

	getCtx, cancel := context.WithTimeout(baseCtx, time.Second*30)
	defer cancel()

	backend, err := clientSet.BackendV1().Get(getCtx, nname)
	if err != nil {
		return err
	}

	var updatedBackend backendv1.Backend
	err = runEditor(backend, &updatedBackend)
	if err != nil {
		return err
	}

	// Exit early if no changes where made
	if proto.Equal(backend, &updatedBackend) {
		logrus.Info("no changes detected")
		os.Exit(0)
	}

	// Send update to server
	updateCtx, cancel := context.WithTimeout(baseCtx, time.Second*30)
	defer cancel()
	err = clientSet.BackendV1().Update(updateCtx, nname, &updatedBackend)
	if err != nil {
		return err
	}

	logrus.Infof("backend %s was updated", nname)

	return nil
}
