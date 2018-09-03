package multikube_test

import (
	"testing"
	"gitlab.com/amimof/multikube"
	"k8s.io/client-go/tools/clientcmd/api"
)
var (
	name string = "minikube"
	defServer string = "https://127.0.0.1:8443"
	defToken string = "aGVsbG93b3JsZA=="
)


var conf *api.Config = &api.Config{
	APIVersion: "v1",
	Kind: "Config",
	Clusters: map[string]*api.Cluster{
		name: &api.Cluster{
			Server: defServer,
		},
	},
	AuthInfos: map[string]*api.AuthInfo{
		name: &api.AuthInfo{
			Token: defToken,
		},
	},
	Contexts: map[string]*api.Context{
		name: &api.Context{
			Cluster: name,
			AuthInfo: name,
		},
	},
	CurrentContext: name,
}

// Just creates a new proxy instance
func TestProxyNewProxy(t *testing.T) {
	p := multikube.NewProxyFrom(conf)
	server := p.Config.Clusters[name].Server
	if server != defServer {
		t.Fatalf("Expected config cluster to be %s, got %s", defServer, server)
	}
	token := p.Config.AuthInfos[name].Token
	if token != defToken {
		t.Fatalf("Expected config token to be %s, got %s", defToken, token)
	}
	context := p.Config.Contexts[name]
	if context.Cluster != name && context.AuthInfo != name {
		t.Fatalf("Expected config context cluster & authinfo to be %s, got cluster: %s authinfo: %s", name, context.Cluster, context.AuthInfo)
	}
	currcontext := p.Config.CurrentContext
	if currcontext != name {
		t.Fatalf("Expected config current-context to be %s, got %s", name, currcontext)
	}
}
