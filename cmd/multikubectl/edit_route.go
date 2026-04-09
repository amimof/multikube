package main

import (
	"context"
	"os"
	"time"

	routev1 "github.com/amimof/multikube/api/route/v1"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"google.golang.org/protobuf/proto"
)

func newEditRouteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "route [NAME]",
		Short: "Edit a route",
		Long:  `Edit a route`,
		Example: `
  # Edit a route
  multikubectl edit route prod-route-v1`,
		Args: cobra.ExactArgs(1),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runEditRouteCmd(cmd, args)
		}),
	}

	return cmd
}

func runEditRouteCmd(cmd *cobra.Command, args []string) error {
	baseCtx := cmd.Context()

	tracer := otel.Tracer("multikubectl")
	baseCtx, span := tracer.Start(baseCtx, "multikubectl.edit.route")
	defer span.End()

	nname := args[0]

	getCtx, routencel := context.WithTimeout(baseCtx, time.Second*30)
	defer routencel()

	route, err := clientSet.RouteV1().Get(getCtx, nname)
	if err != nil {
		return err
	}

	var updatedRoute routev1.Route
	err = runEditor(route, &updatedRoute)
	if err != nil {
		return err
	}

	// Exit early if no changes where made
	if proto.Equal(route, &updatedRoute) {
		logrus.Info("no changes detected")
		os.Exit(0)
	}

	// Send update to server
	updateCtx, routencel := context.WithTimeout(baseCtx, time.Second*30)
	defer routencel()
	err = clientSet.RouteV1().Update(updateCtx, nname, &updatedRoute)
	if err != nil {
		return err
	}

	logrus.Infof("route %s was updated", nname)

	return nil
}
