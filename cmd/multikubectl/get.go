package main

import (
	"github.com/amimof/multikube/pkg/client"
	"github.com/spf13/cobra"
)

func newGetCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get resources",
		Long:  `Get resources`,
		Example: `
# Get all backends
multikube get backends

# Get a specific backend
multikube get backend default-backend

# Get all routes
multikube get routes
`,
		Args: cobra.ExactArgs(1),
	}

	cmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "Output format")

	cmd.AddCommand(newGetBackendCmd(cfg))
	cmd.AddCommand(newGetRouteCmd(cfg))
	cmd.AddCommand(newGetCertificateCmd(cfg))
	cmd.AddCommand(newGetCACmd(cfg))
	cmd.AddCommand(newGetCredentialCmd(cfg))
	cmd.AddCommand(newGetPolicyCmd(cfg))

	return cmd
}
