package main

import (
	"log"
	"github.com/amimof/multikube"
)

func main() {

	m := multikube.New()
	log.Printf("Synchronising cache from %d clusters", len(m.Config.Clusters))
	
	for _, cluster := range m.Config.Clusters {
		err := cluster.Cache().SyncHTTP(cluster)
		if err != nil {
			log.Printf("Error synchronising cluster %s. %s", cluster.Name, err)
		}
	}
	
	log.Printf("Synchronisation complete")
	log.Printf("Listening on http://localhost:8081/\n")

	m.SetupRoutes().ListenAndServe(":8081")

}