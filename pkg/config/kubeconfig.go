package config

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"strings"

	types "github.com/amimof/multikube/api/config/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// GenerateProxyKubeconfig builds a kubeconfig whose server URLs point at the
// multikube proxy rather than directly at Kubernetes API servers. Each backend
// gets a cluster entry with a URL like https://proxy:8443/<pathPrefix>.
//
// If backendNames is empty, all backends that have a matching route are
// included; backends with no route are skipped with a warning to w (stderr is
// expected). If no backends produce output, an error is returned.
//
// The function writes the resulting kubeconfig YAML to out. Warnings (e.g.
// skipped backends) are written to warn. Pass nil for warn to suppress
// warnings.
func GenerateProxyKubeconfig(cfg *types.Config, backendNames []string, out io.Writer, warn io.Writer) error {
	baseURL, err := resolveListenerURL(cfg)
	if err != nil {
		return err
	}

	caData, err := resolveServerCA(cfg)
	if err != nil {
		return fmt.Errorf("resolving server CA: %w", err)
	}

	// Determine the set of backends to process.
	var backends []*types.Backend
	if len(backendNames) > 0 {
		for _, name := range backendNames {
			b := findBackend(cfg, name)
			if b == nil {
				return fmt.Errorf("backend %q not found in config", name)
			}
			backends = append(backends, b)
		}
	} else {
		backends = cfg.Backends
	}

	kubecfg := clientcmdapi.Config{
		APIVersion: "v1",
		Kind:       "Config",
		Clusters:   make(map[string]*clientcmdapi.Cluster),
		AuthInfos:  make(map[string]*clientcmdapi.AuthInfo),
		Contexts:   make(map[string]*clientcmdapi.Context),
	}

	var firstContext string
	for _, backend := range backends {
		route := findRouteForBackend(cfg, backend.Name)
		if route == nil {
			// Explicit backend name → error; all-backends mode → warn and skip.
			if len(backendNames) > 0 {
				return fmt.Errorf("no route found for backend %q", backend.Name)
			}
			if warn != nil {
				fmt.Fprintf(warn, "warning: skipping backend %q: no route found\n", backend.Name)
			}
			continue
		}

		prefix := buildPathPrefix(route, backend.Name)
		serverURL := strings.TrimRight(baseURL, "/") + prefix

		cluster := &clientcmdapi.Cluster{
			Server: serverURL,
		}
		if caData != nil {
			cluster.CertificateAuthorityData = caData
		}

		name := backend.Name
		kubecfg.Clusters[name] = cluster
		kubecfg.AuthInfos[name] = &clientcmdapi.AuthInfo{}
		kubecfg.Contexts[name] = &clientcmdapi.Context{
			Cluster:  name,
			AuthInfo: name,
		}
		if firstContext == "" {
			firstContext = name
		}
	}

	if len(kubecfg.Contexts) == 0 {
		return fmt.Errorf("no backends with routes found")
	}

	kubecfg.CurrentContext = firstContext

	data, err := clientcmd.Write(kubecfg)
	if err != nil {
		return fmt.Errorf("serialising kubeconfig: %w", err)
	}

	_, err = out.Write(data)
	return err
}

// resolveListenerURL determines the proxy server URL from the config.
// It uses the HTTPS listener address if configured, falling back to the
// default "https://localhost:8443". Unix-only configs (no HTTPS listener)
// are a fatal error since kubeconfig cannot represent unix sockets.
func resolveListenerURL(cfg *types.Config) (string, error) {
	if cfg.Server == nil || cfg.Server.Https == nil {
		// No HTTPS listener: check if only unix is configured.
		if cfg.Server != nil && cfg.Server.Unix != nil {
			return "", fmt.Errorf("only unix listener configured; kubeconfig requires an HTTPS listener")
		}
		return defaultListenerURL, nil
	}

	addr := cfg.Server.Https.Address
	if addr == "" {
		return defaultListenerURL, nil
	}

	// The address field uses Go net.Listen format, e.g. ":8443" or "0.0.0.0:8443".
	// We need to produce a URL like "https://localhost:8443".
	host, port, err := parseListenAddress(addr)
	if err != nil {
		return "", fmt.Errorf("parsing https address %q: %w", addr, err)
	}

	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "localhost"
	}

	return fmt.Sprintf("https://%s:%s", host, port), nil
}

// parseListenAddress splits a Go net.Listen-style address (e.g. ":8443",
// "0.0.0.0:8443", "[::]:8443") into host and port components.
func parseListenAddress(addr string) (host, port string, err error) {
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		return "", "", fmt.Errorf("invalid address %q: %w", addr, err)
	}
	return h, p, nil
}

const defaultListenerURL = "https://localhost:8443"

// resolveServerCA returns the PEM-encoded CA certificate for the proxy's
// HTTPS listener, if configured. It reads directly from the inline ca/ca_data
// fields on the HTTPS listener. Returns nil if no CA is set.
func resolveServerCA(cfg *types.Config) ([]byte, error) {
	if cfg.Server == nil || cfg.Server.Https == nil {
		return nil, nil
	}
	h := cfg.Server.Https

	if ca := h.GetCa(); ca != "" {
		// ca is a file path; we don't embed file contents — return nil so the
		// kubeconfig does not include CertificateAuthorityData. The user should
		// set certificate-authority in the kubeconfig manually, or use ca_data.
		return nil, nil
	}

	if caData := h.GetCaData(); caData != "" {
		// Try base64 decode first (kubeconfig-style), fall back to raw PEM.
		decoded, err := base64.StdEncoding.DecodeString(caData)
		if err == nil {
			return decoded, nil
		}
		return []byte(caData), nil
	}

	return nil, nil
}

// findRouteForBackend scans cfg.Routes for the first route whose BackendRef
// matches the given backend name.
func findRouteForBackend(cfg *types.Config, backendName string) *types.Route {
	for _, r := range cfg.Routes {
		if r.BackendRef == backendName {
			return r
		}
	}
	return nil
}

// buildPathPrefix determines the URL path prefix for a given route and backend.
// If the route has an explicit PathPrefix in its match, that is used. Otherwise
// /<backendName> is used as the default.
func buildPathPrefix(route *types.Route, backendName string) string {
	if route.Match != nil && route.Match.PathPrefix != "" {
		p := route.Match.PathPrefix
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		return p
	}
	return "/" + backendName
}
