package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func newConfigDumpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump",
		Short: "View the entire client configuration",
		Long:  "View the entire client configuration",
		Args:  cobra.MaximumNArgs(0),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runConfigDumpCmd()
		}),
	}
	return cmd
}

func runConfigDumpCmd() error {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		logrus.Fatalf("error marshal: %v", err)
	}

	err = os.WriteFile(viper.GetViper().ConfigFileUsed(), b, 0o666)
	if err != nil {
		logrus.Fatalf("error writing config file: %v", err)
	}

	fmt.Println(string(b))

	return nil
}
