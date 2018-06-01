package main

import (
	"log"
	"gitlab.com/amimof/multikube"
	"time"
)

func main() {

	group := multikube.NewGroup("dev").AddCluster(
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
		}
	)
	log.Printf("Performing initial sync")
	
	err := group.SyncAll()
	if err != nil {
		panic(err)
	}
	log.Printf("Cache synced and ready")

	for {
		time.Sleep(time.Second * 5)
		log.Printf("%s", time.Now())
	}

}