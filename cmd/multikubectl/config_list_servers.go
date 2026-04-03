package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newConfigListServersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-servers",
		Short: "List all servers in client configuration",
		Long:  "List all servers in client configuration",
		Args:  cobra.MaximumNArgs(0),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runConfigListServers()
		}),
	}
	return cmd
}

func runConfigListServers() error {
	// Setup writer
	wr := tabwriter.NewWriter(os.Stdout, 8, 8, 8, '\t', tabwriter.AlignRight)

	curr, err := cfg.CurrentServer()
	if err != nil {
		logrus.Fatal(err)
	}

	_, _ = fmt.Fprintf(wr, "%s\t%s\t%s\t%s\n", "NAME", "CURRENT", "ADDRESS", "TLS")

	for _, s := range cfg.Servers {
		isCurrent := curr.Name == s.Name
		hasTLS := s.TLSConfig != nil
		_, _ = fmt.Fprintf(wr, "%s\t%t\t%s\t%t\n",
			s.Name,
			isCurrent,
			s.Address,
			hasTLS,
		)
	}

	_ = wr.Flush()

	return nil
}
