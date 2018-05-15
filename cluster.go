package multikube

import (
	"k8s.io/api/core/v1"
	"github.com/google/uuid"
	"log"
)

type Cluster struct {
	cache *Cache `json:"-"`
	Config *ClusterConfig `json:"config,omitempty"`
}

type ClusterVersion struct {
	BuildDate string `json:"buildDate,omitempty"`
	Compiler string `json:"compiler,omitempty"`
	GitCommit string `json:"gitCommit,omitempty"`
	GitTreeState string `json:"gitTreeState,omitempty"`
	GitVersion string `json:"gitVersion,omitempty"`
	GoVersion string `json:"goVersion,omitempty"`
	Major string `json:"major,omitempty"`
	Minor string `json:"minor,omitempty"`
	Platform string `json:"platform,omitempty"`
}

func (c *Cluster) Version() ClusterVersion {
	version := ClusterVersion{}
	NewRequest(c.Config).Get().Path("/version").Into(&version).Do()
	return version
}

func (c *Cluster) SyncHTTP() *Cache {
	namespaces := &v1.NamespaceList{}
	r, err := NewRequest(c.Config).Get().Namespace("").Into(&namespaces).Do()
	if err != nil {
		log.Printf("WARN: Couldn't sync namspaces for cluster %s, %s", c.Config.Name, err.Error())
	}
	c.Cache().Set("/namespaces/", r.Data())
	return c.Cache()
}

func (c *Cluster) Cache() *Cache {
	if c.cache == nil {
		c.cache = &Cache{
			ID: uuid.New(),
			Store: make(map[string]Item),
		}
	}
	return c.cache
}

func (c *Cluster) Name() string {
	return c.Config.Name
}