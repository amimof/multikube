package multikube

import (
	"log"
	"net/http"
	"encoding/json"
	"k8s.io/api/core/v1"
	"github.com/gorilla/mux"
)

type Cluster struct {
	Name string `json:"name,omitempty"`
	UUID string `json:"uuid,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Token string `json:"token,omitempty"`
	CA string `json:"ca,omitempty"`
	Cert string `json:"cert,omitempty"`
	Key string `json:"key,omitempty"`
	Credential *Credential `json:"credential,omitempty"`
	cache *Cache `json:"-"`
}

func (c *Cluster) Cache() *Cache {
	if c.cache == nil {
		c.cache = &Cache{
			NamespaceList: &v1.NamespaceList{},
			PodList: &v1.PodList{},
			ServiceList: &v1.ServiceList{},
		}
	}
	return c.cache
}

func (c *Cluster) GetNamespaceList() *v1.NamespaceList {
	return c.Cache().NamespaceList
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
	id := mux.Vars(req)["name"]
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

func (m *Multikube) GetClusterNamespacesHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := mux.Vars(req)["name"]
	cluster, err := m.GetCluster(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	data, err := json.Marshal(cluster.Cache().NamespaceList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (m *Multikube) GetClusterNamespaceHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	cluster, err := m.GetCluster(vars["name"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	data, err := json.Marshal(&cluster.Cache().NamespaceList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (m *Multikube) GetClusterPodsHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	cluster, err := m.GetCluster(vars["name"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	data, err := json.Marshal(&cluster.Cache().NamespaceList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (m *Multikube) GetClusterResource(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(req)
	cluster, err := m.GetCluster(vars["name"])
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

func (m *Multikube) GetCluster(name string) (*Cluster, error) {
	for _, cluster := range m.Config.Clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}
	return nil, APIErrorResponse{500, newErrf("Cluster %s does not exist", name)}
}
