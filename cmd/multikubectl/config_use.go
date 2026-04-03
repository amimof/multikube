package main

import (
	"os"

	fzf "github.com/junegunn/fzf/src"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func newConfigUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use NAME",
		Short: "Switch to another server in multikubectl client configuration",
		Long:  "Switch to another server in multikubectl client configuration",
		Example: `
# Switch to server 'production'
multikubectl config use production
`,
		Args: cobra.MaximumNArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runConfigUse(args)
		}),
	}
	return cmd
}

func runConfigUse(args []string) error {
	serverName := cfg.Current

	// Fuzzy finder if no server name is provided on cmd line
	if len(args) == 0 {
		inputChan := make(chan string)
		go func() {
			for _, s := range cfg.Servers {
				inputChan <- s.Name
			}
			close(inputChan)
		}()

		outputChan := make(chan string)
		go func() {
			for s := range outputChan {
				serverName = s
			}
		}()

		options, err := fzf.ParseOptions(
			true,
			nil,
		)
		if err != nil {
			logrus.Fatalf("fzf parse error: %v", err)
		}

		options.Input = inputChan
		options.Output = outputChan

		_, err = fzf.Run(options)
		if err != nil {
			logrus.Fatalf("error running fzf: %v", err)
		}

	}

	// Use server name from args
	if len(args) > 0 {
		serverName = args[0]
		s, err := cfg.GetServer(serverName)
		if err != nil {
			logrus.Fatalf("error using server %s: %v", serverName, err)
		}
		serverName = s.Name
	}

	cfg.Current = serverName

	b, err := yaml.Marshal(cfg)
	if err != nil {
		logrus.Fatalf("error marshal: %v", err)
	}

	err = os.WriteFile(viper.GetViper().ConfigFileUsed(), b, 0o666)
	if err != nil {
		logrus.Fatalf("error writing config file: %v", err)
	}

	logrus.Infof("Using server server %s", serverName)

	return nil
}
