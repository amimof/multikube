package multikube_test

import (
	"gitlab.com/amimof/multikube"
	"k8s.io/client-go/tools/clientcmd/api"
	"testing"
)

var (
	name      string = "minikube"
	defServer string = "https://127.0.0.1:8443"
	defToken  string = "aGVsbG93b3JsZA=="
)

var config *multikube.Config = &multikube.Config{
	OIDCIssuerURL:  "http://localhost:5556/dex",
	RS256PublicKey: nil,
}

var kubeConf *api.Config = &api.Config{
	APIVersion: "v1",
	Kind:       "Config",
	Clusters: map[string]*api.Cluster{
		name: {
			Server: defServer,
		},
	},
	AuthInfos: map[string]*api.AuthInfo{
		name: {
			Token: defToken,
		},
	},
	Contexts: map[string]*api.Context{
		name: {
			Cluster:  name,
			AuthInfo: name,
		},
	},
	CurrentContext: name,
}

// Just creates a new proxy instance
func TestProxyNewProxy(t *testing.T) {
	p := multikube.NewProxyFrom(config, kubeConf)
	server := p.KubeConfig.Clusters[name].Server
	if server != defServer {
		t.Fatalf("Expected config cluster to be %s, got %s", defServer, server)
	}
	token := p.KubeConfig.AuthInfos[name].Token
	if token != defToken {
		t.Fatalf("Expected config token to be %s, got %s", defToken, token)
	}
	context := p.KubeConfig.Contexts[name]
	if context.Cluster != name && context.AuthInfo != name {
		t.Fatalf("Expected config context cluster & authinfo to be %s, got cluster: %s authinfo: %s", name, context.Cluster, context.AuthInfo)
	}
	currcontext := p.KubeConfig.CurrentContext
	if currcontext != name {
		t.Fatalf("Expected config current-context to be %s, got %s", name, currcontext)
	}
}
