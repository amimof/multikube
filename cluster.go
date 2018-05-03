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

func (m *Multikube) GetClustersHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clusters, err := m.GetClusters()
	if err != nil {
		handleErr(w, err)
		return
	}
	data, err := json.Marshal(&clusters)
	if err != nil {
		handleErr(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (m *Multikube) GetClusterHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id, err := strconv.Atoi(strings.TrimPrefix(req.URL.Path, "/clusters/"))

	cluster, err := m.GetCluster(id)
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

func (m *Multikube) CreateClusterHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Create one cluster")
}

func (m *Multikube) UpdateClusterHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Update one cluster")
}

func (m *Multikube) DeleteClusterHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("Delete one cluster")
}

func (m *Multikube) GetClusters() ([]*Cluster, error) {
	return m.Config.Clusters, nil
}

func (m *Multikube) GetCluster(id int) (*Cluster, error) {
	for _, cluster := range m.Config.Clusters {
		if cluster.ID == id {
			return cluster, nil
		}
	}
	return nil, APIErrorResponse{500, newErrf("Cluster with id %d does not exist", id)}
}
