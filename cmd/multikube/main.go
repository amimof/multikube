package main

import (
	"log"
	"net/http"
	"github.com/amimof/multikube"
	"github.com/gorilla/mux"
)

func main() {

	m := multikube.New()
	log.Printf("Synchronising cache from %d clusters", len(m.Config.Clusters))
	for _, cluster := range m.Config.Clusters {
		err := m.Cache.Sync("default", cluster)
		if err != nil {
			log.Printf("Error synchronising cluster %s. %s", cluster.Name, err)
		}
	}

	r := mux.NewRouter()
	r.HandleFunc("/clusters", m.GetClustersHandler).Methods("GET")
	r.HandleFunc("/clusters", m.CreateClusterHandler).Methods("POST")
	r.HandleFunc("/clusters/{id}", m.GetClusterHandler).Methods("GET")
	r.HandleFunc("/clusters/{id}", m.DeleteClusterHandler).Methods("DELETE")
	r.HandleFunc("/clusters/{id}", m.UpdateClusterHandler).Methods("PUT")
	
	r.HandleFunc("/credentials", multikube.GetCredentials).Methods("GET")
	
	log.Printf("Listening on http://localhost:8081/\n")
	http.ListenAndServe(":8081", r)

}