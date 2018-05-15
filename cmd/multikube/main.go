package main

import (
	"log"
	"github.com/amimof/multikube"
)

func main() {

	m := multikube.New()
	log.Printf("Synchronising cache from %d clusters", len(m.Clusters))
	
	for _, cluster := range m.Clusters {
		err := cluster.SyncHTTP()
		if err != nil {
			log.Printf("Error synchronising cluster %s. %s", cluster.Name(), err)
		}
	}
	
	log.Printf("Synchronisation complete")

}