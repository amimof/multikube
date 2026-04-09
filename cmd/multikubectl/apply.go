package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	credentialv1 "github.com/amimof/multikube/api/credential/v1"
	policyv1 "github.com/amimof/multikube/api/policy/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/client/version"
	"github.com/amimof/multikube/pkg/cmdutil"
	"github.com/amimof/multikube/pkg/errs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type genericHeader struct {
	Version string `yaml:"version" json:"version"`
}

type action string

var (
	ActionCreate action = "created"
	ActionUpdate action = "updated"
	ActionDelete action = "deleted"
)

func newApplyCmd() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply resources from a file",
		Long:  "Apply updates or creates resources defined in the provided yaml/json file",
		Example: `
# Apply resources defined as yaml in a file
multikubectl apply -f resources.yaml
`,
		Args: cobra.ExactArgs(0),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			return runApplyCmd(cmd, file)
		}),
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to file with resource definitions")
	if err := cmd.MarkFlagRequired("file"); err != nil {
		logrus.Fatal(err)
	}

	return cmd
}

func runApplyCmd(cmd *cobra.Command, file string) error {
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Read file content
	data, err := os.ReadFile(file)
	if err != nil {
		logrus.Fatalf("error reading file: %v", err)
	}

	// Setup client
	currentSrv, err := cfg.CurrentServer()
	if err != nil {
		logrus.Fatal(err)
	}
	c, err := client.New(currentSrv.Address, client.WithTLSConfigFromCfg(&cfg))
	if err != nil {
		logrus.Fatalf("error setting up client: %v", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			logrus.Fatalf("error closing client connection: %v", err)
		}
	}()

	codec := cmdutil.NewYamlCodec()
	dec := yaml.NewDecoder(strings.NewReader(string(data)))
	idx := 0

	for {

		var raw any
		if err := dec.Decode(&raw); err != nil {
			if err == io.EOF {
				break
			}
			logrus.Fatalf("error decoding document %d: %v", idx, err)
		}

		idx++
		if raw == nil {
			continue
		}

		// Re-encode that doc back to YAML bytes
		docBytes, err := yaml.Marshal(raw)
		if err != nil {
			logrus.Fatalf("error marshal document %d: %v", idx, err)
		}

		// Detect kind by version
		v, err := detectVersion(docBytes)
		if err != nil {
			logrus.Fatalf("error detecting version %d: %v", idx, err)
		}

		switch v {
		case version.VersionBackend:
			var b backendv1.Backend
			if err := codec.Deserialize(docBytes, &b); err != nil {
				logrus.Fatalf("error deserializing %s %d: %v", v, idx, err)
			}
			if _, err := applyBackend(ctx, c, &b, false); err != nil {
				logrus.Fatalf("error applying %s %d: %v", v, idx, err)
			}
		case version.VersionCertificateAuthority:
			var ca cav1.CertificateAuthority
			if err := codec.Deserialize(docBytes, &ca); err != nil {
				logrus.Fatalf("error deserializing %s %d: %v", v, idx, err)
			}
			action, err := applyCA(ctx, c, &ca, true)
			if err != nil {
				return err
			}
			fmt.Printf("certificate authority %q %s\n", ca.GetMeta().GetName(), action)
		case version.VersionCertificate:
			var cert certificatev1.Certificate
			if err := codec.Deserialize(docBytes, &cert); err != nil {
				logrus.Fatalf("error deserializing %s %d: %v", v, idx, err)
			}
			action, err := applyCertificate(ctx, c, &cert, true)
			if err != nil {
				return err
			}
			fmt.Printf("certificate %q %s\n", cert.GetMeta().GetName(), action)
		case version.VersionCredential:
			var cred credentialv1.Credential
			if err := codec.Deserialize(docBytes, &cred); err != nil {
				logrus.Fatalf("error deserializing %s %d: %v", v, idx, err)
			}
			action, err := applyCredential(ctx, c, &cred, true)
			if err != nil {
				return err
			}
			fmt.Printf("credential %q %s\n", cred.GetMeta().GetName(), action)
		case version.VersionPolicy:
			var pol policyv1.Policy
			if err := codec.Deserialize(docBytes, &pol); err != nil {
				logrus.Fatalf("error deserializing %s %d: %v", v, idx, err)
			}
			action, err := applyPolicy(ctx, c, &pol, true)
			if err != nil {
				return err
			}
			fmt.Printf("policy %q %s\n", pol.GetMeta().GetName(), action)
		case version.VersionRoute:
			var route routev1.Route
			if err := codec.Deserialize(docBytes, &route); err != nil {
				logrus.Fatalf("error deserializing %s %d: %v", v, idx, err)
			}
			action, err := applyRoute(ctx, c, &route, true)
			if err != nil {
				return err
			}
			fmt.Printf("route %q %s\n", route.GetMeta().GetName(), action)
		default:
			logrus.Fatalf("document %d: unsupported version %s", idx, v)
		}
	}
	logrus.Infof("applied %d resource(s) from %s", idx, file)
	return nil
}

func detectVersion(doc []byte) (string, error) {
	var h genericHeader
	if err := yaml.Unmarshal(doc, &h); err != nil {
		return "", fmt.Errorf("decode version: %w", err)
	}

	if h.Version == "" {
		return "", fmt.Errorf("missing version field")
	}

	parts := strings.SplitN(h.Version, "/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid version %q", h.Version)
	}
	return h.Version, nil
}

func applyBackend(ctx context.Context, c *client.ClientSet, b *backendv1.Backend, force bool) (action, error) {
	name := b.GetMeta().GetName()
	err := c.BackendV1().Create(ctx, b)
	if err != nil {
		if errs.IsConflict(err) && force {
			if err := c.BackendV1().Update(ctx, name, b); err != nil {
				return "", err
			}
			return ActionUpdate, nil
		}
		return "", err
	}
	return ActionCreate, nil
}

func applyCA(ctx context.Context, c *client.ClientSet, ca *cav1.CertificateAuthority, force bool) (action, error) {
	name := ca.GetMeta().GetName()
	err := c.CAV1().Create(ctx, ca)
	if err != nil {
		if errs.IsConflict(err) && force {
			if err := c.CAV1().Update(ctx, name, ca); err != nil {
				return "", err
			}
			return ActionUpdate, nil
		}
		return "", err
	}
	return ActionCreate, nil
}

func applyCertificate(ctx context.Context, c *client.ClientSet, cert *certificatev1.Certificate, force bool) (action, error) {
	name := cert.GetMeta().GetName()
	err := c.CertificateV1().Create(ctx, cert)
	if err != nil {
		if errs.IsConflict(err) && force {
			if err := c.CertificateV1().Update(ctx, name, cert); err != nil {
				return "", err
			}
			return ActionUpdate, nil
		}
		return "", err
	}
	return ActionCreate, nil
}

func applyCredential(ctx context.Context, c *client.ClientSet, cred *credentialv1.Credential, force bool) (action, error) {
	name := cred.GetMeta().GetName()
	err := c.CredentialV1().Create(ctx, cred)
	if err != nil {
		if errs.IsConflict(err) && force {
			if err := c.CredentialV1().Update(ctx, name, cred); err != nil {
				return "", err
			}
			return ActionUpdate, nil
		}
		return "", err
	}
	return ActionCreate, nil
}

func applyPolicy(ctx context.Context, c *client.ClientSet, policy *policyv1.Policy, force bool) (action, error) {
	name := policy.GetMeta().GetName()
	err := c.PolicyV1().Create(ctx, policy)
	if err != nil {
		if errs.IsConflict(err) && force {
			if err := c.PolicyV1().Update(ctx, name, policy); err != nil {
				return "", err
			}
			return ActionUpdate, nil
		}
		return "", err
	}
	return ActionCreate, nil
}

func applyRoute(ctx context.Context, c *client.ClientSet, route *routev1.Route, force bool) (action, error) {
	name := route.GetMeta().GetName()
	err := c.RouteV1().Create(ctx, route)
	if err != nil {
		if errs.IsConflict(err) && force {
			if err := c.RouteV1().Update(ctx, name, route); err != nil {
				return "", err
			}
			return ActionUpdate, nil
		}
		return "", err
	}
	return ActionCreate, nil
}
