package main

import (
	"log"
	"net/http"
	"github.com/amimof/multikube"
	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/clusters", multikube.GetClustersHandler).Methods("GET")
	r.HandleFunc("/clusters", multikube.CreateClusterHandler).Methods("POST")
	r.HandleFunc("/clusters/{id}", multikube.DeleteClusterHandler).Methods("DELETE")
	r.HandleFunc("/clusters/{id}", multikube.GetClusterHandler).Methods("GET")
	r.HandleFunc("/clusters/{id}", multikube.UpdateClusterHandler).Methods("PUT")
	
	r.HandleFunc("/credentials", multikube.GetCredentials).Methods("GET")
	
	log.Printf("Listening on http://localhost:8081/\n")
	http.ListenAndServe(":8081", r)
}