package main

import (
	"log"
	"gitlab.com/amimof/multikube"
	"time"
)

func main() {

	m := multikube.New()
	log.Printf("Performing initial sync")
	for _, cluster := range m.Clusters {
		cluster.SyncHTTP()
		log.Printf("%s is synced", cluster.Config.Name)
	}

	for {
		time.Sleep(time.Second * 5)
		log.Printf("%s", time.Now())
	}

}