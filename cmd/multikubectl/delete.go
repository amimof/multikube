package main

import (
	"github.com/amimof/multikube/pkg/client"
	"github.com/spf13/cobra"
)

func newDeleteCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources",
		Long:  `Delete resources`,
		Example: `
# Delete a backend
multikube delete backend default-backend

# Delete many backends
multikube delete backend default-backend prod-backend dev-backend
`,
	}

	cmd.AddCommand(newDeleteCertificateCmd(cfg))
	cmd.AddCommand(newDeleteBackendCmd(cfg))
	cmd.AddCommand(newDeleteCertificateAuthorityCmd(cfg))

	return cmd
}
