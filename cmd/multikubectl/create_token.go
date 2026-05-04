package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
	"github.com/amimof/multikube/pkg/client"
	tokenclientv1 "github.com/amimof/multikube/pkg/client/token/v1"
)

var createTokenClientFactory = func() tokenclientv1.ClientV1 {
	return clientSet.TokenV1()
}

func newCreateTokenCmd(cfg *client.Config) *cobra.Command {
	var (
		subject      string
		username     string
		groups       []string
		serviceAccts []string
		audience     string
		scopes       []string
		clusters     []string
		ttlSeconds   uint64
		extraClaims  []string
	)

	cmd := &cobra.Command{
		Use:   "token",
		Short: "Create a new token",
		Long:  `Create a new token using the server signing key.`,
		Example: `  multikubectl create token --subject alice
  multikubectl create token --subject alice --group platform --service-account default/builder --claim team=platform`,
		Args: cobra.ExactArgs(0),
		RunE: withClientSet(func(cmd *cobra.Command, args []string) error {
			return runCreateTokenCmd(cmd, cfg, subject, username, audience, groups, serviceAccts, scopes, clusters, ttlSeconds, extraClaims)
		}),
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Subject claim to issue as `sub`")
	cmd.Flags().StringVar(&username, "username", "", "Optional preferred username claim")
	cmd.Flags().StringVar(&audience, "audience", "", "Audience claim value, repeatable")
	cmd.Flags().StringArrayVar(&groups, "group", []string{}, "Group claim value, repeatable")
	cmd.Flags().StringArrayVar(&serviceAccts, "service-account", []string{}, "Service account claim value, repeatable")
	cmd.Flags().StringArrayVar(&scopes, "scope", []string{}, "Optional scope values, repeatable")
	cmd.Flags().StringArrayVar(&clusters, "cluster", []string{}, "Optional cluster values, repeatable")
	cmd.Flags().Uint64Var(&ttlSeconds, "ttl", 0, "Token time-to-live in seconds. Set to 0 to make the access token never expire")
	cmd.Flags().StringArrayVar(&extraClaims, "claim", []string{}, "Additional claim in key=value format, repeatable")
	_ = cmd.MarkFlagRequired("subject")

	return cmd
}

func claimMapFromStringArray(extraClaims []string) (map[string]string, error) {
	if extraClaims == nil {
		return nil, fmt.Errorf("extraClaims is nil")
	}
	res := map[string]string{}
	for _, claim := range extraClaims {
		split := strings.Split(claim, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("%s is not a valid format", claim)
		}
		res[split[0]] = split[1]
	}
	return res, nil
}

// runCreateCreateCmd creates a new route via the multikube API server.
func runCreateTokenCmd(
	cmd *cobra.Command,
	cfg *client.Config,
	subject, username, audience string,
	groups, serviceAccounts, scopes, clusters []string,
	ttlSeconds uint64,
	extraClaims []string,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.token.create")
	defer span.End()

	extraClaimsMap, err := claimMapFromStringArray(extraClaims)
	if err != nil {
		logrus.Fatalf("%v", err)
	}

	req := &tokenv1.Token{
		Config: &tokenv1.TokenConfig{
			Subject:         subject,
			Username:        username,
			Groups:          groups,
			ServiceAccounts: serviceAccounts,
			Audience:        audience,
			Scopes:          scopes,
			Clusters:        clusters,
			ExtraClaims:     extraClaimsMap,
		},
	}
	if cmd.Flags().Changed("ttl") {
		req.Config.Ttl = &ttlSeconds
	}

	res, err := createTokenClientFactory().IssueToken(ctx, req)
	if err != nil {
		logrus.Fatalf("error creating token: %v", err)
	}

	_, err = fmt.Fprintln(os.Stdout, res.GetAccessToken())
	if err != nil {
		return err
	}

	return nil
}
