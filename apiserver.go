package multikube

import (
	"k8s.io/api/core/v1"
	//"k8s.io/api/apps/v1beta1"
	"github.com/google/uuid"

	apimachineryversion "k8s.io/apimachinery/pkg/version"

	"encoding/json"
	"path"
	//"log"
)

type ResourceSpec struct {
	ApiVersion string
	Name       string
	Namespace  string
	Type       interface{}
	Path       string
}

type APIServer struct {
	Name     string `json:"name,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	CA       string `json:"ca,omitempty"`
	Cert     string `json:"cert,omitempty"`
	Key      string `json:"key,omitempty"`
	cache		 *Cache
}

// Version returns the version of the connected backend.
// This does a call to /version on the Kubernetes API.
func (c *APIServer) Version() apimachineryversion.Info {
	version := apimachineryversion.Info{}
	NewRequest(fromAPIServer(c)).Get().Path("/version").Into(&version).Do()
	return version
}

// SyncHTTP performs a full syncronisation of the cluster and it's configured endpoint.
// A cache is instantiated if none is available. Any items that are stored in the clusters cache
// will be overwritten.
func (c *APIServer) SyncHTTP() (*Cache, error) {

	opt := fromAPIServer(c)

	// Sync namespaces
	nslist := v1.NamespaceList{}
	r, err := NewRequest(opt).Get().Namespace("/").Into(&nslist).Do()
	if err != nil {
		return c.Cache(), err
	}
	c.Cache().Set("/namespaces/", r.Data())

	// Sync Nodes
	//nodelist := v1.NodeList{}
	r, err = NewRequest(opt).Get().Resource("Nodes").Do()
	if err != nil {
		return c.Cache(), err
	}
	c.Cache().Set("/nodes", r.Data())

	// Sync PersistentVolumes
	//pvlist := v1.PersistentVolumeList{}
	r, err = NewRequest(opt).Get().Resource("PersistentVolumes").Do()
	if err != nil {
		return c.Cache(), err
	}
	c.Cache().Set("/persistentvolumes", r.Data())

	resources := []ResourceSpec{
		ResourceSpec{ApiVersion: "v1", Name: "pods"},
		ResourceSpec{ApiVersion: "v1", Name: "services"},
		ResourceSpec{ApiVersion: "v1", Name: "serviceaccounts"},
		ResourceSpec{ApiVersion: "v1", Name: "secrets"},
		ResourceSpec{ApiVersion: "v1", Name: "replicationcontrollers"},
		ResourceSpec{ApiVersion: "v1", Name: "configmaps"},
		ResourceSpec{ApiVersion: "v1", Name: "events"},
		ResourceSpec{ApiVersion: "v1", Name: "limitranges"},
		ResourceSpec{ApiVersion: "v1", Name: "persistentvolumeclaims"},
		ResourceSpec{ApiVersion: "v1", Name: "podtemplates"},
		ResourceSpec{ApiVersion: "v1", Name: "resourcequotas"},
		ResourceSpec{ApiVersion: "apps/v1beta1", Name: "deployments"},
		ResourceSpec{ApiVersion: "apps/v1beta1", Name: "statefulsets"},
	}

	// Sync namespace and its resources
	for _, ns := range nslist.Items {

		r, err := NewRequest(opt).Get().Namespace(ns.ObjectMeta.Name).Do()
		if err != nil {
			return c.Cache(), nil
		}
		c.Cache().Set(path.Join("/namespaces/", ns.ObjectMeta.Name), r.Data())

		for _, resource := range resources {
			r, err := NewRequest(opt).Get().Resource(resource.Name).Namespace(ns.ObjectMeta.Name).ApiVer(resource.ApiVersion).Do()
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
func (c *APIServer) Namespaces() v1.NamespaceList {
	data := c.Cache().Get("/namespaces/").Value
	nslist := v1.NamespaceList{}
	_ = json.Unmarshal(data.([]byte), &nslist)
	return nslist
}

// Namespace return a Namespace by name of cluster avaialble in its cache. Returns nil if error
func (c *APIServer) Namespace(name string) v1.Namespace {
	data := c.Cache().Get(path.Join("/namespaces/", name)).Value
	ns := v1.Namespace{}
	_ = json.Unmarshal(data.([]byte), &ns)
	return ns
}

// Cache returns the current cache instance of the cluster
func (c *APIServer) Cache() *Cache {
	if c.cache == nil {
		c.cache = &Cache{
			ID:    uuid.New(),
			Store: make(map[string]Item),
		}
	}
	return c.cache
}

func fromAPIServer(a *APIServer) *Options {
	return &Options{
		Hostname: a.Hostname,
		CA: a.CA,
		Cert: a.Cert,
		Key: a.Key,
	}
}