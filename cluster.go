package multikube

import (
	"k8s.io/api/core/v1"
	"k8s.io/api/apps/v1beta1"
	"github.com/google/uuid"
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

type ResourceSpec struct {
	ApiVersion string
	Name string
	Namespace string
	Type interface{}
	Path string
}


func (c *Cluster) Version() ClusterVersion {
	version := ClusterVersion{}
	NewRequest(c.Config).Get().Path("/version").Into(&version).Do()
	return version
}

func (c *Cluster) SyncHTTP() (*Cache, error) {
	
	// Sync namespaces
	nslist := v1.NamespaceList{}
	r, err := NewRequest(c.Config).Get().Namespace("/").Into(&nslist).Do()
	if err != nil {
		return c.Cache(), err
	}
	c.Cache().Set("/namespaces/", r.Data())

	// Sync Nodes
	nodelist := v1.NodeList{}
	r, err = NewRequest(c.Config).Get().Resource("Nodes").Into(&nodelist).Do()
	if err != nil {
		return c.Cache(), err
	}
	c.Cache().Set("/nodes", r.Data())
	
	// Sync PersistentVolumes
	pvlist := v1.PersistentVolumeList{}
	r, err = NewRequest(c.Config).Get().Resource("PersistentVolumes").Into(&pvlist).Do()
	if err != nil {
		return c.Cache(), err
	}
	c.Cache().Set("/persistentvolumes", r.Data())

	resources := []ResourceSpec{
		ResourceSpec{ ApiVersion: "v1", Name: "pods", Type: &v1.PodList{}, Path: "/namespaces/pods/" },
		ResourceSpec{ ApiVersion: "v1", Name: "services", Type: &v1.ServiceList{}, Path: "/namespaces/services/" },
		ResourceSpec{ ApiVersion: "v1", Name: "serviceaccounts", Type: &v1.ServiceAccountList{}, Path: "/namespaces/serviceaccounts/" },
		ResourceSpec{ ApiVersion: "v1", Name: "secrets", Type: &v1.SecretList{}, Path: "/namespaces/secrets" },
		ResourceSpec{ ApiVersion: "v1", Name: "replicationcontrollers", Type: &v1.ReplicationControllerList{}, Path: "/namespaces/replicationcontrollers" },
		ResourceSpec{ ApiVersion: "v1", Name: "configmaps", Type: &v1.ConfigMapList{}, Path: "/namespaces/configmaps" },
		ResourceSpec{ ApiVersion: "v1", Name: "events", Type: &v1.EventList{}, Path: "/namespaces/events" },
		ResourceSpec{ ApiVersion: "v1", Name: "limitranges", Type: &v1.LimitRangeList{}, Path: "/namespaces/limitranges" },
		ResourceSpec{ ApiVersion: "v1", Name: "persistentvolumeclaims", Type: &v1.PersistentVolumeClaimList{}, Path: "/namespaces/persistentvolumeclaims" },
		ResourceSpec{ ApiVersion: "v1", Name: "podtemplates", Type: &v1.PodTemplateList{}, Path: "/namespaces/podtemplates" },
		ResourceSpec{ ApiVersion: "v1", Name: "resourcequotas", Type: &v1.ResourceQuotaList{}, Path: "/namespaces/resourcequotas" },
		ResourceSpec{ ApiVersion: "apps/v1beta1", Name: "deployments", Type: &v1beta1.DeploymentList{}, Path: "/namespaces/deployments", },
		ResourceSpec{ ApiVersion: "apps/v1beta1", Name: "statefulsets", Type: &v1beta1.StatefulSetList{}, Path: "/namespaces/statefulsets" },
	}

	// Sync namespace resources
	for _, ns := range nslist.Items {

		for _, resource := range resources {
			r, err := NewRequest(c.Config).Get().Resource(resource.Name).Namespace(ns.ObjectMeta.Name).ApiVer(resource.ApiVersion).Into(resource.Type).Do()
			if err != nil {
				return c.Cache(), nil
			}
			c.Cache().Set(resource.Path, r.Data())
		}

	}
	return c.Cache(), nil
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