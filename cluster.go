package multikube

import (
	"log"
	"net/http"
	"encoding/json"
	"strings"
	"strconv"
)

type Cluster struct {
	Name string `json:"name,omitempty"`
	ID int `json:"id,omitempty"`
	UUID string `json:"uuid,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Credential *Credential `json:"credential,omitempty"`
}

var clusters []*Cluster = []*Cluster{
	&Cluster{
		ID: 1,
		Name: "cluster01",
		UUID: "f4098feb-3e84-49c1-a862-7864b15049a4",
	},
	&Cluster{
		ID: 2,
		Name: "cluster02",
		UUID: "f906d9d1-0090-4c24-abdd-d5bb3b5ee2ad",
	},
}

func GetClustersHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clusters, err := GetClusters()
	if err != nil {
		handleResponse(w, err)
		return
	}
	data, err := json.Marshal(&clusters)
	if err != nil {
		handleResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func GetClusterHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(strings.TrimPrefix(req.URL.Path, "/clusters/"))

	cluster, err := GetCluster(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	data, err := json.Marshal(&cluster)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func GetClusters() ([]*Cluster, error) {
	return clusters, nil
}

func GetCluster(id int) (*Cluster, error) {
	for _, cluster := range clusters {
		if cluster.ID == id {
			return cluster, nil
		}
	}
	return nil, APIErrorResponse{500, newErrf("Cluster with id %d does not exist", id)}
}

func CreateClusterHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Create one cluster")
}

func UpdateClusterHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Update one cluster")
}

func DeleteClusterHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Delete one cluster")
}
