package config

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"

	types "github.com/amimof/multikube/api/config/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ExportKubeconfig builds a kubeconfig from a named backend in the multikube
// config and writes it to w. The generated kubeconfigcontains exactly one
// cluster, one user, and one context.
//
// References (ca_ref, auth_ref, client_certificate_ref) are resolved within
// the provided config. File-path references (certificate, key, ca certificate)
// are read from disk so that the output kubeconfig is self-contained with
// inline data.
func ExportKubeconfig(cfg *types.Config, backendName string, w io.Writer) error {
	// Find the backend.
	backend := findBackend(cfg, backendName)
	if backend == nil {
		return fmt.Errorf("backend %q not found in config", backendName)
	}

	kubecfg := clientcmdapi.Config{
		APIVersion:     "v1",
		Kind:           "Config",
		Clusters:       make(map[string]*clientcmdapi.Cluster),
		AuthInfos:      make(map[string]*clientcmdapi.AuthInfo),
		Contexts:       make(map[string]*clientcmdapi.Context),
		CurrentContext: backendName,
	}

	if err := exportBackendInto(cfg, backend, &kubecfg); err != nil {
		return err
	}

	data, err := clientcmd.Write(kubecfg)
	if err != nil {
		return fmt.Errorf("serialising kubeconfig: %w", err)
	}

	_, err = w.Write(data)
	return err
}

// ExportAllKubeconfigs builds a single kubeconfig containing every backend in
// the multikube config and writes it to w. Each backend produces one cluster,
// one user, and one context. The first backend
// is set as current-context.
func ExportAllKubeconfigs(cfg *types.Config, w io.Writer) error {
	if len(cfg.Backends) == 0 {
		return fmt.Errorf("no backends defined in config")
	}

	kubecfg := clientcmdapi.Config{
		APIVersion:     "v1",
		Kind:           "Config",
		Clusters:       make(map[string]*clientcmdapi.Cluster),
		AuthInfos:      make(map[string]*clientcmdapi.AuthInfo),
		Contexts:       make(map[string]*clientcmdapi.Context),
		CurrentContext: cfg.Backends[0].Name,
	}

	for _, backend := range cfg.Backends {
		if err := exportBackendInto(cfg, backend, &kubecfg); err != nil {
			return err
		}
	}

	data, err := clientcmd.Write(kubecfg)
	if err != nil {
		return fmt.Errorf("serialising kubeconfig: %w", err)
	}

	_, err = w.Write(data)
	return err
}

// exportBackendInto resolves a single backend and adds its cluster, authinfo,
// and context entries into the given kubeconfig.
func exportBackendInto(cfg *types.Config, backend *types.Backend, kubecfg *clientcmdapi.Config) error {
	name := backend.Name

	cluster := &clientcmdapi.Cluster{
		Server:                backend.Server,
		InsecureSkipTLSVerify: backend.InsecureSkipTlsVerify,
	}

	// Resolve CA.
	if backend.CaRef != "" {
		ca := findCertificateAuthority(cfg, backend.CaRef)
		if ca == nil {
			return fmt.Errorf("backend %q: ca_ref %q not found in config", name, backend.CaRef)
		}
		caData, err := resolveCAData(ca)
		if err != nil {
			return fmt.Errorf("backend %q: resolving CA %q: %w", name, ca.Name, err)
		}
		cluster.CertificateAuthorityData = caData
	}

	authInfo := &clientcmdapi.AuthInfo{}

	// Resolve credential.
	if backend.AuthRef != "" {
		cred := findCredential(cfg, backend.AuthRef)
		if cred == nil {
			return fmt.Errorf("backend %q: auth_ref %q not found in config", name, backend.AuthRef)
		}
		if err := populateAuthInfo(cfg, name, cred, authInfo); err != nil {
			return err
		}
	}

	kubecfg.Clusters[name] = cluster
	kubecfg.AuthInfos[name] = authInfo
	kubecfg.Contexts[name] = &clientcmdapi.Context{
		Cluster:  name,
		AuthInfo: name,
	}

	return nil
}

// populateAuthInfo fills in the kubeconfig AuthInfo from a multikube Credential.
func populateAuthInfo(cfg *types.Config, backendName string, cred *types.Credential, authInfo *clientcmdapi.AuthInfo) error {
	switch {
	case cred.ClientCertificateRef != "":
		cert := findCertificate(cfg, cred.ClientCertificateRef)
		if cert == nil {
			return fmt.Errorf("backend %q: credential %q references unknown certificate %q",
				backendName, cred.Name, cred.ClientCertificateRef)
		}
		certData, err := resolveCertData(cert)
		if err != nil {
			return fmt.Errorf("backend %q: resolving certificate %q: %w", backendName, cert.Name, err)
		}
		keyData, err := resolveKeyData(cert)
		if err != nil {
			return fmt.Errorf("backend %q: resolving key for certificate %q: %w", backendName, cert.Name, err)
		}
		authInfo.ClientCertificateData = certData
		authInfo.ClientKeyData = keyData

	case cred.Token != "":
		authInfo.Token = cred.Token

	case cred.Basic != nil:
		authInfo.Username = cred.Basic.Username
		authInfo.Password = cred.Basic.Password
	}

	return nil
}

// resolveCAData returns raw PEM bytes for a CertificateAuthority, reading from
// file or decoding the inline data as needed.
func resolveCAData(ca *types.CertificateAuthority) ([]byte, error) {
	if ca.Certificate != "" {
		return os.ReadFile(ca.Certificate)
	}
	if ca.CertificateData != "" {
		// The import command stores CA data as base64(PEM). Try base64
		// decode first; if that fails, treat as raw PEM.
		decoded, err := base64.StdEncoding.DecodeString(ca.CertificateData)
		if err == nil {
			return decoded, nil
		}
		return []byte(ca.CertificateData), nil
	}
	return nil, fmt.Errorf("no certificate data")
}

// resolveCertData returns raw PEM bytes for the certificate portion of a
// Certificate entry.
func resolveCertData(cert *types.Certificate) ([]byte, error) {
	if cert.Certificate != "" {
		return os.ReadFile(cert.Certificate)
	}
	if cert.CertificateData != "" {
		return []byte(cert.CertificateData), nil
	}
	return nil, fmt.Errorf("no certificate data")
}

// resolveKeyData returns raw PEM bytes for the key portion of a Certificate
// entry.
func resolveKeyData(cert *types.Certificate) ([]byte, error) {
	if cert.Key != "" {
		return os.ReadFile(cert.Key)
	}
	if cert.KeyData != "" {
		return []byte(cert.KeyData), nil
	}
	return nil, fmt.Errorf("no key data")
}

func findBackend(cfg *types.Config, name string) *types.Backend {
	for _, b := range cfg.Backends {
		if b.Name == name {
			return b
		}
	}
	return nil
}

func findCertificateAuthority(cfg *types.Config, name string) *types.CertificateAuthority {
	for _, ca := range cfg.CertificateAuthorities {
		if ca.Name == name {
			return ca
		}
	}
	return nil
}

func findCredential(cfg *types.Config, name string) *types.Credential {
	for _, c := range cfg.Credentials {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func findCertificate(cfg *types.Config, name string) *types.Certificate {
	for _, c := range cfg.Certificates {
		if c.Name == name {
			return c
		}
	}
	return nil
}
