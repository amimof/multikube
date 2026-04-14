package main

import (
	"context"
	"fmt"
	"time"

	credentialv1 "github.com/amimof/multikube/api/credential/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/cmdutil"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

func newCreateCredentialCmd(cfg *client.Config) *cobra.Command {
	var (
		token          string
		basicUsername  string
		basicPassword  string
		certificateRef string
		labels         []string
	)

	cmd := &cobra.Command{
		Use:   "credential [NAME]",
		Short: "Create a new credential",
		Long:  `Create a new credential and register it with the server.`,
		Args:  cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runCreateCredentialCmd(cmd, args, cfg, token, basicUsername, basicPassword, certificateRef, labels)
		}),
	}

	cmd.Flags().StringVar(&token, "token", "", "Bearer token to use for upstream authentication")
	cmd.Flags().StringVar(&basicUsername, "basic-username", "", "Username for upstream basic authentication")
	cmd.Flags().StringVar(&basicPassword, "basic-password", "", "Password for upstream basic authentication")
	cmd.Flags().StringVar(&certificateRef, "certificate-ref", "", "Reference to a certificate resource for upstream mTLS")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Labels to attach in key=value format (can be specified multiple times)")

	return cmd
}

func runCreateCredentialCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
	token, basicUsername, basicPassword, certificateRef string,
	labelStrs []string,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.credential.create")
	defer span.End()

	currentSrv, err := cfg.CurrentServer()
	if err != nil {
		logrus.Fatal(err)
	}

	c, err := client.New(currentSrv.Address, client.WithTLSConfigFromCfg(cfg))
	if err != nil {
		logrus.Fatalf("error setting up client: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			logrus.Errorf("error closing client connection: %v", err)
		}
	}()

	credentialConfig, err := buildCredentialConfig(args[0], token, basicUsername, basicPassword, certificateRef)
	if err != nil {
		return err
	}

	credential := &credentialv1.Credential{
		Meta: &metav1.Meta{
			Name:   args[0],
			Labels: cmdutil.ConvertKVStringsToMap(labelStrs),
		},
		Config: credentialConfig,
	}

	if err := c.CredentialV1().Create(ctx, credential); err != nil {
		logrus.Fatalf("error creating credential: %v", err)
	}

	fmt.Printf("credential %q created\n", args[0])

	return nil
}

func buildCredentialConfig(name, token, basicUsername, basicPassword, certificateRef string) (*credentialv1.CredentialConfig, error) {
	methodCount := 0
	if token != "" {
		methodCount++
	}
	if certificateRef != "" {
		methodCount++
	}
	if basicUsername != "" || basicPassword != "" {
		if basicUsername == "" || basicPassword == "" {
			return nil, fmt.Errorf("--basic-username and --basic-password must both be set")
		}
		methodCount++
	}

	if methodCount != 1 {
		return nil, fmt.Errorf("exactly one of --token, --certificate-ref, or basic auth flags must be set")
	}

	config := &credentialv1.CredentialConfig{}

	switch {
	case token != "":
		config.Token = token
	case certificateRef != "":
		config.ClientCertificateRef = certificateRef
	default:
		config.Basic = &credentialv1.CredentialBasic{
			Username: basicUsername,
			Password: basicPassword,
		}
	}

	return config, nil
}
