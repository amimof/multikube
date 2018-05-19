package multikube_test

import (
	"gitlab.com/amimof/multikube"
)

var configs []*multikube.Config = []*multikube.Config{
	&multikube.Config{
		Name: "minikube",
    Hostname: "https://192.168.99.100:8443",
    Cert: "/Users/amir/.minikube/client.crt",
    Key: "/Users/amir/.minikube/client.key",
    CA: "/Users/amir/.minikube/ca.crt",
	},
	&multikube.Config{
		Name: "prod-cluster-1",
    Hostname: "https://192.168.99.100:8443",
    Cert: "/Users/amir/.minikube/client.crt",
    Key: "/Users/amir/.minikube/client.key",
    CA: "/Users/amir/.minikube/ca.crt",
	},
}

var group *multikube.Group = multikube.NewGroup("test").AddCluster(configs...)