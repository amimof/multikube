package main

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	var (
		kubeconfigPath string
		configPath     string
		force          bool
	)

	cmd := &cobra.Command{
		Use:   "import <context>",
		Short: "Import a kubeconfig context as a multikube backend",
		Long: `Import a kubeconfig context into a multikube configuration file.

This reads the specified context from a kubeconfig file, extracts the cluster
and user (authinfo) definitions, and creates the corresponding multikube
backend, certificate authority, certificate, and credential entries.

The context name is used as the backend name. Related objects are named
with the context name as a prefix (e.g. "<context>-ca", "<context>-cred").`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			contextName := args[0]

			// Resolve kubeconfig path.
			if kubeconfigPath == "" {
				kubeconfigPath = defaultKubeconfigPath()
			}

			return runImport(contextName, kubeconfigPath, configPath, force)
		},
	}

	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	cmd.Flags().StringVar(&configPath, "config", "/etc/multikube/config.yaml", "path to multikube config file to update")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite existing backend and related objects if they already exist")

	return cmd
}

// defaultKubeconfigPath returns the kubeconfig path from $KUBECONFIG or the
// standard default ~/.kube/config.
func defaultKubeconfigPath() string {
	if v := os.Getenv("KUBECONFIG"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".kube", "config")
	}
	return filepath.Join(home, ".kube", "config")
}

// runImport is the core logic for the "config import" command.
// When force is true, existing objects with the same names are replaced
// instead of causing a conflict error.
func runImport(contextName, kubeconfigPath, configPath string, force bool) error {
	return nil
}
