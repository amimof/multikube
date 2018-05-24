package multikube

import (
	"k8s.io/api/core/v1"
	//"k8s.io/api/apps/v1beta1"
	"github.com/google/uuid"
	"path"
	"encoding/json"
)

type Cluster struct {
	cache *Cache `json:"-"`
	Config *Config `json:"config,omitempty"`
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

type ResourceSpec struct {
	ApiVersion string
	Name string
	Namespace string
	Type interface{}
	Path string
}

// Version returns the version of the connected backend. 
// This does a call to /version on the Kubernetes API.
func (c *Cluster) Version() ClusterVersion {
	version := ClusterVersion{}
	NewRequest(c.Config).Get().Path("/version").Into(&version).Do()
	return version
}

// SyncHTTP performs a full syncronisation of the cluster and it's configured endpoint. 
// A cache is instantiated if none is available. Any items that are stored in the clusters cache
// will be overwritten.
func (c *Cluster) SyncHTTP() (*Cache, error) {
	
	// Sync namespaces
	nslist := v1.NamespaceList{}
	r, err := NewRequest(c.Config).Get().Namespace("/").Into(&nslist).Do()
	if err != nil {
		return c.Cache(), err
	}
	c.Cache().Set("/namespaces/", r.Data())

	// Sync Nodes
	//nodelist := v1.NodeList{}
	r, err = NewRequest(c.Config).Get().Resource("Nodes").Do()
	if err != nil {
		return c.Cache(), err
	}
	c.Cache().Set("/nodes", r.Data())
	
	// Sync PersistentVolumes
	//pvlist := v1.PersistentVolumeList{}
	r, err = NewRequest(c.Config).Get().Resource("PersistentVolumes").Do()
	if err != nil {
		return c.Cache(), err
	}
	c.Cache().Set("/persistentvolumes", r.Data())

	resources := []ResourceSpec{
		ResourceSpec{ ApiVersion: "v1", Name: "pods" },
		ResourceSpec{ ApiVersion: "v1", Name: "services" },
		ResourceSpec{ ApiVersion: "v1", Name: "serviceaccounts" },
		ResourceSpec{ ApiVersion: "v1", Name: "secrets" },
		ResourceSpec{ ApiVersion: "v1", Name: "replicationcontrollers" },
		ResourceSpec{ ApiVersion: "v1", Name: "configmaps" },
		ResourceSpec{ ApiVersion: "v1", Name: "events" },
		ResourceSpec{ ApiVersion: "v1", Name: "limitranges" },
		ResourceSpec{ ApiVersion: "v1", Name: "persistentvolumeclaims" },
		ResourceSpec{ ApiVersion: "v1", Name: "podtemplates" },
		ResourceSpec{ ApiVersion: "v1", Name: "resourcequotas" },
		ResourceSpec{ ApiVersion: "apps/v1beta1", Name: "deployments" },
		ResourceSpec{ ApiVersion: "apps/v1beta1", Name: "statefulsets" },
	}

	// Sync namespace and its resources
	for _, ns := range nslist.Items {
		
		r, err := NewRequest(c.Config).Get().Namespace(ns.ObjectMeta.Name).Do()
		if err != nil {
			return c.Cache(), nil
		} 
		c.Cache().Set(path.Join("/namespaces/", ns.ObjectMeta.Name), r.Data())

		for _, resource := range resources {
			r, err := NewRequest(c.Config).Get().Resource(resource.Name).Namespace(ns.ObjectMeta.Name).ApiVer(resource.ApiVersion).Do()
			if err != nil {
				return c.Cache(), nil
			}
			p := path.Join("/namespaces/", ns.ObjectMeta.Name, resource.Name)
			c.Cache().Set(p, r.Data())
		}

	}
	return c.Cache(), nil
}

// Namespaces returns NamespaceList of cluster available in its cache. Returns nil if error
func (c *Cluster) Namespaces() v1.NamespaceList {
	data := c.Cache().Get("/namespaces/").Value
	nslist := v1.NamespaceList{}
	_ = json.Unmarshal(data.([]byte), &nslist)
	return nslist
}

// Namespace return a Namespace by name of cluster avaialble in its cache. Returns nil if error
func (c *Cluster) Namespace(name string) v1.Namespace {
	data := c.Cache().Get(path.Join("/namespaces/", name)).Value
	ns := v1.Namespace{}
	_ = json.Unmarshal(data.([]byte), &ns)
	return ns
}

// Cache returns the current cache instance of the cluster
func (c *Cluster) Cache() *Cache {
	if c.cache == nil {
		c.cache = &Cache{
			ID: uuid.New(),
			Store: make(map[string]Item),
		}
	}
	return c.cache
}