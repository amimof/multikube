package main

import (
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage multikube configuration",
		Long:  "Commands for viewing and modifying multikube configuration files.",
	}

	cmd.AddCommand(newExportCmd())
	cmd.AddCommand(newInitCmd())

	return cmd
}
