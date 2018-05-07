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
		err := cluster.Cache().SyncHTTP(cluster)
		if err != nil {
			log.Printf("Error synchronising cluster %s. %s", cluster.Name, err)
		}
	}
	log.Printf("Synchronisation complete")

	r := mux.NewRouter()

	// Cluster
	r.HandleFunc("/clusters", m.GetClustersHandler).Methods("GET")
	r.HandleFunc("/clusters", m.CreateClusterHandler).Methods("POST")
	r.HandleFunc("/clusters/{name}", m.GetClusterHandler).Methods("GET")
	r.HandleFunc("/clusters/{name}", m.DeleteClusterHandler).Methods("DELETE")
	r.HandleFunc("/clusters/{name}", m.UpdateClusterHandler).Methods("PUT")
	
	// Namespace
	r.HandleFunc("/clusters/{name}/namespaces", m.GetClusterNamespacesHandler).Methods("GET")
	r.HandleFunc("/clusters/{name}/namespaces/{ns}", m.GetClusterNamespaceHandler).Methods("GET")
	
	// Resource
	r.HandleFunc("/clusters/{name}/namespaces/{ns}/{resource}", m.GetClusterResource).Methods("GET")

	log.Printf("Listening on http://localhost:8081/\n")
	http.ListenAndServe(":8081", r)

}