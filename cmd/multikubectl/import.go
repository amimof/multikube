package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	types "github.com/amimof/multikube/api/config/v1"
	mkconfig "github.com/amimof/multikube/pkg/config"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func newImportCmd() *cobra.Command {
	var (
		kubeconfigPath string
		configPath     string
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

			return runImport(contextName, kubeconfigPath, configPath)
		},
	}

	cmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	cmd.Flags().StringVar(&configPath, "config", "/etc/multikube/config.yaml", "path to multikube config file to update")

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
func runImport(contextName, kubeconfigPath, configPath string) error {
	// 1. Load the kubeconfig.
	kubecfg, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("loading kubeconfig %s: %w", kubeconfigPath, err)
	}

	// 2. Look up the context.
	ctx, ok := kubecfg.Contexts[contextName]
	if !ok {
		return fmt.Errorf("context %q not found in kubeconfig %s", contextName, kubeconfigPath)
	}

	cluster := kubecfg.Clusters[ctx.Cluster]
	if cluster == nil {
		return fmt.Errorf("cluster %q (referenced by context %q) not found in kubeconfig", ctx.Cluster, contextName)
	}

	authInfo := kubecfg.AuthInfos[ctx.AuthInfo]
	// authInfo may be nil — some contexts have no user.

	// 3. Build proto objects from the kubeconfig data.
	result, err := buildImportObjects(contextName, cluster, authInfo)
	if err != nil {
		return err
	}

	// 4. Load or create the multikube config.
	cfg, err := loadOrCreateConfig(configPath)
	if err != nil {
		return err
	}

	// 5. Check for name conflicts.
	if err := checkNameConflicts(cfg, result); err != nil {
		return err
	}

	// 6. Append new objects and write.
	appendObjects(cfg, result)

	if err := writeConfigFile(cfg, configPath); err != nil {
		return err
	}

	fmt.Printf("Imported context %q into %s\n", contextName, configPath)
	printImportSummary(result)

	return nil
}

// importResult holds the proto objects generated from a kubeconfig context.
type importResult struct {
	Backend              *types.Backend
	CertificateAuthority *types.CertificateAuthority // may be nil
	Certificate          *types.Certificate          // client cert, may be nil
	Credential           *types.Credential           // may be nil
}

// buildImportObjects converts a kubeconfig context (cluster + authinfo) into
// multikube proto objects.
func buildImportObjects(contextName string, cluster *api.Cluster, authInfo *api.AuthInfo) (*importResult, error) {
	result := &importResult{}

	// --- Backend ---
	if cluster.Server == "" {
		return nil, fmt.Errorf("cluster for context %q has no server URL", contextName)
	}
	backend := &types.Backend{
		Name:                  contextName,
		Server:                cluster.Server,
		InsecureSkipTlsVerify: cluster.InsecureSkipTLSVerify,
	}
	result.Backend = backend

	// --- Certificate Authority ---
	ca := buildCertificateAuthority(contextName, cluster)
	if ca != nil {
		result.CertificateAuthority = ca
		backend.CaRef = ca.Name
	}

	// --- Client Certificate + Credential ---
	if authInfo != nil {
		cert := buildClientCertificate(contextName, authInfo)
		if cert != nil {
			result.Certificate = cert
		}

		cred := buildCredential(contextName, authInfo, cert)
		if cred != nil {
			result.Credential = cred
			backend.AuthRef = cred.Name
		}
	}

	return result, nil
}

// buildCertificateAuthority creates a CertificateAuthority proto from cluster
// CA data. Returns nil if the cluster has no CA configuration.
func buildCertificateAuthority(contextName string, cluster *api.Cluster) *types.CertificateAuthority {
	name := contextName + "-ca"

	if len(cluster.CertificateAuthorityData) > 0 {
		// Inline CA data — store as base64-encoded PEM.
		// The Convert() function tries base64 decode first.
		encoded := base64.StdEncoding.EncodeToString(cluster.CertificateAuthorityData)
		return &types.CertificateAuthority{
			Name:            name,
			CertificateData: encoded,
		}
	}

	if cluster.CertificateAuthority != "" {
		// File path reference.
		return &types.CertificateAuthority{
			Name:        name,
			Certificate: cluster.CertificateAuthority,
		}
	}

	return nil
}

// buildClientCertificate creates a Certificate proto from authinfo client cert
// data. Returns nil if no client certificate is configured.
func buildClientCertificate(contextName string, authInfo *api.AuthInfo) *types.Certificate {
	name := contextName + "-cert"

	if len(authInfo.ClientCertificateData) > 0 && len(authInfo.ClientKeyData) > 0 {
		// Inline cert/key — store as raw PEM strings.
		// The Convert() function uses them directly as PEM bytes.
		return &types.Certificate{
			Name:            name,
			CertificateData: string(authInfo.ClientCertificateData),
			KeyData:         string(authInfo.ClientKeyData),
		}
	}

	if authInfo.ClientCertificate != "" && authInfo.ClientKey != "" {
		// File path references.
		return &types.Certificate{
			Name:        name,
			Certificate: authInfo.ClientCertificate,
			Key:         authInfo.ClientKey,
		}
	}

	return nil
}

// buildCredential creates a Credential proto from authinfo data.
// Returns nil if no authentication method is configured.
func buildCredential(contextName string, authInfo *api.AuthInfo, clientCert *types.Certificate) *types.Credential {
	name := contextName + "-cred"

	// Priority: client certificate > bearer token > basic auth.
	// This matches kubeconfig precedence.

	if clientCert != nil {
		return &types.Credential{
			Name:                 contextName + "-cred",
			ClientCertificateRef: clientCert.Name,
		}
	}

	if authInfo.Token != "" {
		return &types.Credential{
			Name:  name,
			Token: authInfo.Token,
		}
	}

	if authInfo.Username != "" && authInfo.Password != "" {
		return &types.Credential{
			Name: name,
			Basic: &types.CredentialBasic{
				Username: authInfo.Username,
				Password: authInfo.Password,
			},
		}
	}

	return nil
}

// loadOrCreateConfig loads an existing multikube config file, or returns an
// empty Config if the file does not exist.
func loadOrCreateConfig(path string) (*types.Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &types.Config{}, nil
	}
	cfg, err := mkconfig.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading existing config %s: %w", path, err)
	}
	return cfg, nil
}

// checkNameConflicts verifies that none of the new objects collide with
// existing entries in the config.
func checkNameConflicts(cfg *types.Config, r *importResult) error {
	// Check backend name.
	for _, b := range cfg.Backends {
		if b.Name == r.Backend.Name {
			return fmt.Errorf("backend %q already exists in config", r.Backend.Name)
		}
	}

	// Check CA name.
	if r.CertificateAuthority != nil {
		for _, ca := range cfg.CertificateAuthorities {
			if ca.Name == r.CertificateAuthority.Name {
				return fmt.Errorf("certificate authority %q already exists in config", r.CertificateAuthority.Name)
			}
		}
	}

	// Check certificate name.
	if r.Certificate != nil {
		for _, c := range cfg.Certificates {
			if c.Name == r.Certificate.Name {
				return fmt.Errorf("certificate %q already exists in config", r.Certificate.Name)
			}
		}
	}

	// Check credential name.
	if r.Credential != nil {
		for _, c := range cfg.Credentials {
			if c.Name == r.Credential.Name {
				return fmt.Errorf("credential %q already exists in config", r.Credential.Name)
			}
		}
	}

	return nil
}

// appendObjects adds the imported objects to the config.
func appendObjects(cfg *types.Config, r *importResult) {
	if r.CertificateAuthority != nil {
		cfg.CertificateAuthorities = append(cfg.CertificateAuthorities, r.CertificateAuthority)
	}
	if r.Certificate != nil {
		cfg.Certificates = append(cfg.Certificates, r.Certificate)
	}
	if r.Credential != nil {
		cfg.Credentials = append(cfg.Credentials, r.Credential)
	}
	cfg.Backends = append(cfg.Backends, r.Backend)
}

// writeConfigFile marshals the config to YAML and writes it atomically to the
// given path (write to temp file in the same directory, then rename).
func writeConfigFile(cfg *types.Config, path string) error {
	data, err := mkconfig.MarshalYAML(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	dir := filepath.Dir(path)

	// Ensure the directory exists.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config directory %s: %w", dir, err)
	}

	// Write to a temp file in the same directory for atomic rename.
	tmp, err := os.CreateTemp(dir, ".multikube-config-*.yaml")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file to %s: %w", path, err)
	}

	return nil
}

// printImportSummary prints a summary of what was imported.
// TODO: Use tablewriter here
func printImportSummary(r *importResult) {
	fmt.Printf("  Backend:               %s (server: %s)\n", r.Backend.Name, r.Backend.Server)
	if r.CertificateAuthority != nil {
		source := "inline"
		if r.CertificateAuthority.Certificate != "" {
			source = "file: " + r.CertificateAuthority.Certificate
		}
		fmt.Printf("  Certificate Authority: %s (%s)\n", r.CertificateAuthority.Name, source)
	}
	if r.Certificate != nil {
		source := "inline"
		if r.Certificate.Certificate != "" {
			source = "file: " + r.Certificate.Certificate
		}
		fmt.Printf("  Certificate:           %s (%s)\n", r.Certificate.Name, source)
	}
	if r.Credential != nil {
		method := "unknown"
		if r.Credential.ClientCertificateRef != "" {
			method = "client certificate"
		} else if r.Credential.Token != "" {
			method = "bearer token"
		} else if r.Credential.Basic != nil {
			method = "basic auth"
		}
		fmt.Printf("  Credential:            %s (%s)\n", r.Credential.Name, method)
	}
}
