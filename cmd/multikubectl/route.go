package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"

	metav1 "github.com/amimof/multikube/api/meta/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/cmdutil"
)

func newRouteCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "route",
		Aliases: []string{"routes"},
		Short:   "Manage routes",
		Long:    `Manage routes`,
	}
	cmd.AddCommand(newRouteListCmd(cfg))
	cmd.AddCommand(newRouteCreateCmd(cfg))

	return cmd
}

func newRouteListCmd(cfg *client.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List routes registered with the multikube API server",
		Long: `Retrieve and display all routes currently registered with
the multikube API server.

The command connects to the API server whose address is resolved from the
current server entry in the multikube configuration file and prints a
summary table with each route's name, generation, and age.`,
		Example: `  # List all routes using the default configuration
  multikubectl route list

  # List routes using a custom configuration file
  multikubectl route list --config /etc/multikube/config.yaml`,
		Args: cobra.NoArgs,
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runRouteListCmd(cmd, cfg)
		}),
	}

	return cmd
}

func newRouteCreateCmd(cfg *client.Config) *cobra.Command {
	var (
		backendRef  string
		pathPrefix  string
		sni         string
		headerName  string
		headerValue string
		jwtClaim    string
		jwtValue    string
		labels      []string
	)

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new route",
		Long: `Create a new route and register it with the multikube API server.

The NAME argument is required and sets the route's name. Use --backend-ref to
associate the route with a backend. Match criteria can be set via --sni,
--path-prefix, --header-name/--header-value, or --jwt-claim/--jwt-value.`,
		Example: `  # Create a route pointing to a backend
  multikubectl route create my-route --backend-ref my-cluster

  # Create a route with SNI matching
  multikubectl route create my-route --backend-ref my-cluster --sni api.example.com

  # Create a route with path prefix matching
  multikubectl route create my-route --backend-ref my-cluster --path-prefix /api/v1

  # Create a route with header matching
  multikubectl route create my-route --backend-ref my-cluster \
    --header-name X-Tenant --header-value acme

  # Create a route with JWT claim matching
  multikubectl route create my-route --backend-ref my-cluster \
    --jwt-claim tenant --jwt-value acme

  # Create a route with labels
  multikubectl route create my-route --backend-ref my-cluster \
    --label env=production --label team=platform`,
		Args: cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runRouteCreateCmd(cmd, args, cfg, backendRef, pathPrefix, sni, headerName, headerValue, jwtClaim, jwtValue, labels)
		}),
	}

	cmd.Flags().StringVar(&backendRef, "backend-ref", "", "Reference to the backend this route targets")
	cmd.Flags().StringVar(&pathPrefix, "path-prefix", "", "Path prefix to match on incoming requests")
	cmd.Flags().StringVar(&sni, "sni", "", "Server Name Indication (SNI) value to match")
	cmd.Flags().StringVar(&headerName, "header-name", "", "HTTP header name to match")
	cmd.Flags().StringVar(&headerValue, "header-value", "", "HTTP header value to match (used together with --header-name)")
	cmd.Flags().StringVar(&jwtClaim, "jwt-claim", "", "JWT claim name to match")
	cmd.Flags().StringVar(&jwtValue, "jwt-value", "", "JWT claim value to match (used together with --jwt-claim)")
	cmd.Flags().StringArrayVar(&labels, "label", nil, "Labels to attach in key=value format (can be specified multiple times)")

	return cmd
}

// runRouteCreateCmd creates a new route via the multikube API server.
func runRouteCreateCmd(
	cmd *cobra.Command,
	args []string,
	cfg *client.Config,
	backendRef, pathPrefix, sni, headerName, headerValue, jwtClaim, jwtValue string,
	labelStrs []string,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.route.create")
	defer span.End()

	name := args[0]

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

	match := &routev1.Match{
		Sni:        sni,
		PathPrefix: pathPrefix,
	}

	if headerName != "" || headerValue != "" {
		match.Header = &routev1.HeaderMatch{
			Name:  headerName,
			Value: headerValue,
		}
	}

	if jwtClaim != "" || jwtValue != "" {
		match.Jwt = &routev1.JWTMatch{
			Claim: jwtClaim,
			Value: jwtValue,
		}
	}

	route := &routev1.Route{
		Meta: &metav1.Meta{
			Name:   name,
			Labels: cmdutil.ConvertKVStringsToMap(labelStrs),
		},
		Config: &routev1.RouteConfig{
			Name:       name,
			BackendRef: backendRef,
			Match:      match,
		},
	}

	if err := c.RouteV1().Create(ctx, route); err != nil {
		logrus.Fatalf("error creating route: %v", err)
	}

	fmt.Printf("route %q created\n", name)

	return nil
}

// runRouteListCmd lists all routes registered with the multikube API server
// and prints them as a formatted table to stdout.
func runRouteListCmd(cmd *cobra.Command, cfg *client.Config) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second*30)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.route.list")
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

	wr := tabwriter.NewWriter(os.Stdout, 8, 8, 8, '\t', tabwriter.AlignRight)

	routes, err := c.RouteV1().List(ctx)
	if err != nil {
		logrus.Fatal(err)
	}

	_, _ = fmt.Fprintf(wr, "%s\t%s\t%s\n", "NAME", "GENERATION", "AGE")
	for _, route := range routes {
		_, _ = fmt.Fprintf(wr, "%s\t%d\t%s\n",
			route.GetMeta().GetName(),
			route.GetMeta().GetGeneration(),
			cmdutil.FormatDuration(time.Since(route.GetMeta().GetCreated().AsTime())),
		)
	}

	_ = wr.Flush()

	return nil
}
