package main

import (
	"log"
	"gitlab.com/amimof/multikube"
	"time"
)

func main() {

	cluster1 := &multikube.ClusterConfig{
		Name: "minikube",
    Hostname: "https://192.168.99.100:8443",
    Cert: "/Users/amir/.minikube/client.crt",
    Key: "/Users/amir/.minikube/client.key",
    CA: "/Users/amir/.minikube/ca.crt",
	}
	cluster2 := &multikube.ClusterConfig{
		Name: "prod-cluster-1",
    Hostname: "https://192.168.99.100:8443",
    Cert: "/Users/amir/.minikube/client.crt",
    Key: "/Users/amir/.minikube/client.key",
    CA: "/Users/amir/.minikube/ca.crt",
	}

	g := multikube.NewGroup("dev").AddCluster(cluster1, cluster2)
	log.Printf("Performing initial sync")
	
	for _, cluster := range g.Clusters() {
		cache, err := cluster.SyncHTTP()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s is synced. Current cache size: %d bytes", cluster.Config.Name, cache.Size())
	}

	for {
		time.Sleep(time.Second * 5)
		log.Printf("%s", time.Now())
	}

}