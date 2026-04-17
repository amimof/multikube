package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	credentialv1 "github.com/amimof/multikube/api/credential/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/errs"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type importResourceNames struct {
	Backend              string
	Credential           string
	Certificate          string
	CertificateAuthority string
}

type importPlan struct {
	Names       importResourceNames
	Backend     *backendv1.Backend
	Credential  *credentialv1.Credential
	Certificate *certificatev1.Certificate
	CA          *cav1.CertificateAuthority
}

func newImportCmd(cfg *client.Config) *cobra.Command {
	var (
		kubeconfigPath           string
		backendName              string
		credentialName           string
		certificateName          string
		certificateAuthorityName string
		force                    bool
	)

	cmd := &cobra.Command{
		Use:   "import <context>",
		Short: "Import a kubeconfig context into multikube",
		Long: `Import a kubeconfig context multikube.

This reads the specified context from a kubeconfig file, extracts the cluster
and user (authinfo) definitions, and creates the corresponding multikube
backend, certificate authority, certificate, and credential entries.

		Related resources are named from the kubeconfig context by default:
		"<context>-backend", "<context>-credential", "<context>-certificate",
		and "<context>-certificate-authority".

		If a target resource already exists the import fails fast. Use --force to
		update existing resources instead.`,
		Args: cobra.ExactArgs(1),
		RunE: withConfig(func(cmd *cobra.Command, args []string) error {
			contextName := args[0]

			if kubeconfigPath == "" {
				kubeconfigPath = defaultKubeconfigPath()
			}

			return runImport(cmd, cfg, contextName, kubeconfigPath, force, importResourceNames{
				Backend:              backendName,
				Credential:           credentialName,
				Certificate:          certificateName,
				CertificateAuthority: certificateAuthorityName,
			})
		}),
	}

	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	cmd.Flags().StringVar(&backendName, "backend-name", "", "name for the imported backend (default: <context>-backend)")
	cmd.Flags().StringVar(&credentialName, "credential-name", "", "name for the imported credential (default: <context>-credential)")
	cmd.Flags().StringVar(&certificateName, "certificate-name", "", "name for the imported certificate (default: <context>-certificate)")
	cmd.Flags().StringVar(&certificateAuthorityName, "certificate-authority-name", "", "name for the imported certificate authority (default: <context>-certificate-authority)")
	cmd.Flags().BoolVar(&force, "force", false, "update existing resources instead of failing when they already exist")

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

func defaultImportResourceNames(contextName string) importResourceNames {
	return importResourceNames{
		Backend:              contextName + "-backend",
		Credential:           contextName + "-credential",
		Certificate:          contextName + "-certificate",
		CertificateAuthority: contextName + "-certificate-authority",
	}
}

func (n importResourceNames) withDefaults(contextName string) importResourceNames {
	defaults := defaultImportResourceNames(contextName)
	if n.Backend == "" {
		n.Backend = defaults.Backend
	}
	if n.Credential == "" {
		n.Credential = defaults.Credential
	}
	if n.Certificate == "" {
		n.Certificate = defaults.Certificate
	}
	if n.CertificateAuthority == "" {
		n.CertificateAuthority = defaults.CertificateAuthority
	}
	return n
}

func runImport(
	cmd *cobra.Command,
	cfg *client.Config,
	contextName, kubeconfigPath string,
	force bool,
	names importResourceNames,
) error {
	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	tracer := otel.Tracer("multikubectl")
	ctx, span := tracer.Start(ctx, "multikubectl.import")
	defer span.End()

	kubeconfig, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("error loading kubeconfig %q: %w", kubeconfigPath, err)
	}

	plan, err := buildImportPlan(kubeconfigPath, contextName, kubeconfig, names)
	if err != nil {
		return err
	}

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

	if !force {
		if err := preflightImportPlan(ctx, c, plan); err != nil {
			return err
		}
	}

	if plan.CA != nil {
		action, err := applyCA(ctx, c, plan.CA, force)
		if err != nil {
			return err
		}
		fmt.Printf("certificateauthority %q %s\n", plan.CA.GetMeta().GetName(), action)
	}

	if plan.Certificate != nil {
		action, err := applyCertificate(ctx, c, plan.Certificate, force)
		if err != nil {
			return err
		}
		fmt.Printf("certificate %q %s\n", plan.Certificate.GetMeta().GetName(), action)
	}

	if plan.Credential != nil {
		action, err := applyCredential(ctx, c, plan.Credential, force)
		if err != nil {
			return err
		}
		fmt.Printf("credential %q %s\n", plan.Credential.GetMeta().GetName(), action)
	}

	action, err := applyBackend(ctx, c, plan.Backend, force)
	if err != nil {
		return err
	}
	fmt.Printf("backend %q %s\n", plan.Backend.GetMeta().GetName(), action)

	return nil
}

func buildImportPlan(kubeconfigPath, contextName string, kubeconfig *clientcmdapi.Config, names importResourceNames) (*importPlan, error) {
	if kubeconfig == nil {
		return nil, fmt.Errorf("kubeconfig is nil")
	}

	ctxDef, ok := kubeconfig.Contexts[contextName]
	if !ok {
		return nil, fmt.Errorf("context %q not found in kubeconfig", contextName)
	}
	if ctxDef == nil {
		return nil, fmt.Errorf("context %q is nil", contextName)
	}
	if ctxDef.Cluster == "" {
		return nil, fmt.Errorf("context %q does not reference a cluster", contextName)
	}

	cluster, ok := kubeconfig.Clusters[ctxDef.Cluster]
	if !ok {
		return nil, fmt.Errorf("cluster %q referenced by context %q not found in kubeconfig", ctxDef.Cluster, contextName)
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster %q referenced by context %q is nil", ctxDef.Cluster, contextName)
	}

	var authInfo *clientcmdapi.AuthInfo
	if ctxDef.AuthInfo != "" {
		var found bool
		authInfo, found = kubeconfig.AuthInfos[ctxDef.AuthInfo]
		if !found {
			return nil, fmt.Errorf("authinfo %q referenced by context %q not found in kubeconfig", ctxDef.AuthInfo, contextName)
		}
		if authInfo == nil {
			return nil, fmt.Errorf("authinfo %q referenced by context %q is nil", ctxDef.AuthInfo, contextName)
		}
	}

	names = names.withDefaults(contextName)
	plan := &importPlan{Names: names}

	caPEM, err := readKubeconfigContent(kubeconfigPath, cluster.CertificateAuthorityData, cluster.CertificateAuthority, "certificate authority")
	if err != nil {
		return nil, err
	}
	if caPEM != "" {
		plan.CA = &cav1.CertificateAuthority{
			Meta: &metav1.Meta{Name: names.CertificateAuthority},
			Config: &cav1.CertificateAuthorityConfig{
				CertificateData: caPEM,
			},
		}
	}

	credentialConfig, certificateObj, err := buildImportedAuth(kubeconfigPath, authInfo, names)
	if err != nil {
		return nil, err
	}
	if certificateObj != nil {
		plan.Certificate = certificateObj
	}
	if credentialConfig != nil {
		plan.Credential = &credentialv1.Credential{
			Meta:   &metav1.Meta{Name: names.Credential},
			Config: credentialConfig,
		}
	}

	backend := &backendv1.Backend{
		Meta: &metav1.Meta{Name: names.Backend},
		Config: &backendv1.BackendConfig{
			Servers:               []string{cluster.Server},
			InsecureSkipTlsVerify: cluster.InsecureSkipTLSVerify,
		},
	}
	if plan.CA != nil {
		backend.Config.CaRef = names.CertificateAuthority
	}
	if plan.Credential != nil {
		backend.Config.AuthRef = names.Credential
	}
	plan.Backend = backend

	return plan, nil
}

func buildImportedAuth(kubeconfigPath string, authInfo *clientcmdapi.AuthInfo, names importResourceNames) (*credentialv1.CredentialConfig, *certificatev1.Certificate, error) {
	if authInfo == nil {
		return nil, nil, nil
	}
	if authInfo.AuthProvider != nil {
		return nil, nil, fmt.Errorf("unsupported auth method: auth-provider")
	}
	if authInfo.Exec != nil {
		return nil, nil, fmt.Errorf("unsupported auth method: exec")
	}

	token, err := readToken(kubeconfigPath, authInfo.Token, authInfo.TokenFile)
	if err != nil {
		return nil, nil, err
	}

	hasToken := token != ""
	hasBasic := authInfo.Username != "" || authInfo.Password != ""
	hasClientCertMaterial := authInfo.ClientCertificate != "" || len(authInfo.ClientCertificateData) > 0 || authInfo.ClientKey != "" || len(authInfo.ClientKeyData) > 0

	if hasBasic && (authInfo.Username == "" || authInfo.Password == "") {
		return nil, nil, fmt.Errorf("basic auth requires both username and password")
	}
	if hasClientCertMaterial {
		hasCert := authInfo.ClientCertificate != "" || len(authInfo.ClientCertificateData) > 0
		hasKey := authInfo.ClientKey != "" || len(authInfo.ClientKeyData) > 0
		if hasCert != hasKey {
			return nil, nil, fmt.Errorf("client certificate auth requires both certificate and key")
		}
	}

	methodCount := 0
	if hasToken {
		methodCount++
	}
	if hasBasic {
		methodCount++
	}
	if hasClientCertMaterial {
		methodCount++
	}
	if methodCount > 1 {
		return nil, nil, fmt.Errorf("kubeconfig authinfo contains multiple supported auth methods; choose one")
	}
	if methodCount == 0 {
		return nil, nil, nil
	}

	config := &credentialv1.CredentialConfig{}
	if hasToken {
		config.Token = token
		return config, nil, nil
	}
	if hasBasic {
		config.Basic = &credentialv1.CredentialBasic{
			Username: authInfo.Username,
			Password: authInfo.Password,
		}
		return config, nil, nil
	}

	certificatePEM, err := readKubeconfigContent(kubeconfigPath, authInfo.ClientCertificateData, authInfo.ClientCertificate, "client certificate")
	if err != nil {
		return nil, nil, err
	}
	keyPEM, err := readKubeconfigContent(kubeconfigPath, authInfo.ClientKeyData, authInfo.ClientKey, "client key")
	if err != nil {
		return nil, nil, err
	}
	certificateObj := &certificatev1.Certificate{
		Meta: &metav1.Meta{Name: names.Certificate},
		Config: &certificatev1.CertificateConfig{
			CertificateData: certificatePEM,
			KeyData:         keyPEM,
		},
	}
	config.ClientCertificateRef = names.Certificate
	return config, certificateObj, nil
}

func readToken(kubeconfigPath, token, tokenFile string) (string, error) {
	if token != "" {
		return strings.TrimSpace(token), nil
	}
	if tokenFile == "" {
		return "", nil
	}
	b, err := os.ReadFile(resolveKubeconfigPath(kubeconfigPath, tokenFile))
	if err != nil {
		return "", fmt.Errorf("error reading token file %q: %w", tokenFile, err)
	}
	return strings.TrimSpace(string(b)), nil
}

func readKubeconfigContent(kubeconfigPath string, inline []byte, refPath, kind string) (string, error) {
	if len(inline) > 0 {
		return string(inline), nil
	}
	if refPath == "" {
		return "", nil
	}
	b, err := os.ReadFile(resolveKubeconfigPath(kubeconfigPath, refPath))
	if err != nil {
		return "", fmt.Errorf("error reading %s file %q: %w", kind, refPath, err)
	}
	return string(b), nil
}

func resolveKubeconfigPath(kubeconfigPath, value string) string {
	if value == "" || filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(filepath.Dir(kubeconfigPath), value)
}

func preflightImportPlan(ctx context.Context, c *client.ClientSet, plan *importPlan) error {
	if plan.CA != nil {
		if err := ensureResourceMissing("certificateauthority", plan.CA.GetMeta().GetName(), func() error {
			_, err := c.CAV1().Get(ctx, plan.CA.GetMeta().GetName())
			return err
		}); err != nil {
			return err
		}
	}
	if plan.Certificate != nil {
		if err := ensureResourceMissing("certificate", plan.Certificate.GetMeta().GetName(), func() error {
			_, err := c.CertificateV1().Get(ctx, plan.Certificate.GetMeta().GetName())
			return err
		}); err != nil {
			return err
		}
	}
	if plan.Credential != nil {
		if err := ensureResourceMissing("credential", plan.Credential.GetMeta().GetName(), func() error {
			_, err := c.CredentialV1().Get(ctx, plan.Credential.GetMeta().GetName())
			return err
		}); err != nil {
			return err
		}
	}
	return ensureResourceMissing("backend", plan.Backend.GetMeta().GetName(), func() error {
		_, err := c.BackendV1().Get(ctx, plan.Backend.GetMeta().GetName())
		return err
	})
}

func ensureResourceMissing(kind, name string, get func() error) error {
	err := get()
	if err == nil {
		return fmt.Errorf("%s %q already exists", kind, name)
	}
	if errs.IsNotFound(err) {
		return nil
	}
	return fmt.Errorf("error checking existing %s %q: %w", kind, name, err)
}
