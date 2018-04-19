package main

import (
	"log"
	"github.com/amimof/multikube"
)

func main() {

	router := multikube.NewRouter()
	router.HandleFunc("/clusters", multikube.GetClustersHandler)
	router.HandleFunc("/clusters/1", multikube.GetClusterHandler)
	
	router.HandleFunc("/credentials", multikube.GetCredentials)
	
	log.Printf("Listening on http://localhost:8081/\n")
	router.Listen(":8081")
}