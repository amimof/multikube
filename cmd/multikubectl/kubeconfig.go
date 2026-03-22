package main

import (
	"github.com/spf13/cobra"
)

func newKubeconfigCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "kubeconfig [backend]",
		Short: "Generate a kubeconfig that targets the multikube proxy",
		Long: `Generate a kubeconfig whose server URLs point at the multikube proxy
rather than directly at Kubernetes API servers.

Each backend that has a matching route gets a context entry. The server URL
is constructed from the proxy's listener address and the route's path prefix.

If a backend name is given, only that backend is included. Without an argument
all backends with routes are included; backends without routes are skipped
with a warning.

Examples:

  multikubectl kubeconfig staging > ~/.kube/multikube-staging.yaml
  multikubectl kubeconfig > ~/.kube/multikube-all.yaml`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var backendNames []string
			if len(args) == 1 {
				backendNames = []string{args[0]}
			}
			return runKubeconfig(backendNames, configPath)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "/etc/multikube/config.yaml", "path to multikube config file to read")

	return cmd
}

// runKubeconfig generates a proxy-targeting kubeconfig and writes it to stdout.
func runKubeconfig(backendNames []string, configPath string) error {
	return nil
}
