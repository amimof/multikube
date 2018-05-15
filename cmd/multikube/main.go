package main

import (
	"log"
	"github.com/amimof/multikube"
)

func main() {

	m := multikube.New()
	
	log.Printf("%+v", m.Clusters)	

	for _, cluster := range m.Clusters {
		log.Printf("%+v", cluster.Config)
	}


}