package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// VERSION of the app. Set at build time via -ldflags.
	VERSION string
	// COMMIT is the Git commit. Set at build time via -ldflags.
	COMMIT string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "multikubectl",
		Short: "CLI for managing multikube configuration",
		Long:  "multikubectl is a command-line tool for managing multikube configuration files.",
	}

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newConfigCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\nCommit: %s\n", VERSION, COMMIT)
		},
	}
}
