package multikube

import (
	//"k8s.io/api/core/v1"
	"github.com/google/uuid"
)

type Cluster struct {
	cache *Cache `json:"-"`
	Config *ClusterConfig `json:"config,omitempty"`
}

func (c *Cluster) Cache() *Cache {
	if c.cache == nil {
		c.cache = &Cache{
			ID: uuid.New(),
		}
	}
	return c.cache
}

func (c *Cluster) Name() string {
	return c.Config.Name
}