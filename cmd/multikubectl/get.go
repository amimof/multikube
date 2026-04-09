package main

import (
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
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

	cmd.AddCommand(newGetBackendCmd())
	cmd.AddCommand(newGetRouteCmd())
	cmd.AddCommand(newGetCertificateCmd())
	cmd.AddCommand(newGetCACmd())
	cmd.AddCommand(newGetCredentialCmd())
	cmd.AddCommand(newGetPolicyCmd())

	return cmd
}
