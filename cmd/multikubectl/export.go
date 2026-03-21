package main

import (
	"fmt"
	"os"

	mkconfig "github.com/amimof/multikube/pkg/config"
	"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "export [backend]",
		Short: "Export multikube backends to kubeconfig format",
		Long: `Export one or all multikube backends as kubeconfig YAML on stdout.

If a backend name is given, only that backend is exported. Without an argument
all backends are exported into a single kubeconfig.

Pipe the output to a file or merge it into an existing kubeconfig as needed:

  multikubectl config export my-backend > ~/.kube/my-backend.yaml
  multikubectl config export > ~/.kube/all-backends.yaml
  KUBECONFIG=~/.kube/config:~/.kube/my-backend.yaml kubectl config view --flatten`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return runExport(args[0], configPath)
			}
			return runExportAll(configPath)
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "/etc/multikube/config.yaml", "path to multikube config file to read")

	return cmd
}

// runExport exports a single backend as kubeconfig YAML to stdout.
func runExport(backendName, configPath string) error {
	cfg, err := mkconfig.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("loading multikube config %s: %w", configPath, err)
	}

	return mkconfig.ExportKubeconfig(cfg, backendName, os.Stdout)
}

// runExportAll exports all backends as a single kubeconfig YAML to stdout.
func runExportAll(configPath string) error {
	cfg, err := mkconfig.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("loading multikube config %s: %w", configPath, err)
	}

	return mkconfig.ExportAllKubeconfigs(cfg, os.Stdout)
}
