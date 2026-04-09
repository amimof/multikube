package main

import (
	"github.com/amimof/multikube/pkg/client"
	"github.com/spf13/cobra"
)

func newCreateCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create resources",
		Long:  `Create resources`,
		Example: `
# Create all backends
multikube create backends

# Create a specific backend
multikube create backend default-backend

# Create all routes
multikube create routes
`,
		Args: cobra.ExactArgs(1),
	}

	cmd.AddCommand(newCreateBackendCmd(cfg))
	cmd.AddCommand(newCreateRouteCmd(cfg))
	cmd.AddCommand(newCreateCertificateCmd(cfg))
	cmd.AddCommand(newCreateCACmd(cfg))
	cmd.AddCommand(newCreateCredentialCmd(cfg))
	cmd.AddCommand(newCreatePolicyCmd(cfg))
	cmd.AddCommand(newCreateTokenCmd(cfg))

	return cmd
}
