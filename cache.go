package multikube

import (
	"k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/tools/clientcmd"
	// "k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/tools/cache"
	//"k8s.io/apimachinery/pkg/watch"
	//"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/fields"
	"log"
)

// Root cache object
type Cache struct {
	NamespaceList *v1.NamespaceList `json:"namespaces"`
	NamespaceResources map[string]NamespaceResource
	client *Request
}

// Namespaced resources
type NamespaceResource struct {
	PodList *v1.PodList `json:"pods,omitempty"`
	ServiceList *v1.ServiceList `json:"services,omitempty"`
}

func (c *Cache) SyncHTTP(cl *Cluster) error {
	
	nsls := &v1.NamespaceList{}
	r, err := NewRequest(cl).Get().Resource("namespaces").Into(nsls).Do()
	if err != nil {
		return err
	}
	c.NamespaceList = nsls
	log.Printf("Found %d namespaces", len(c.NamespaceList.Items))

	for _, namespace := range nsls.Items {
		podls := &v1.PodList{}
		r, err = r.Get().Resource("Pods").Namespace(namespace.ObjectMeta.Name).Into(podls).Do()
		if err != nil {
			return err
		}

		svcls := &v1.ServiceList{}
		r, err = r.Get().Resource("Services").Namespace(namespace.ObjectMeta.Name).Into(svcls).Do()
		if err != nil {
			return err
		}

		c.NamespaceResources[namespace.ObjectMeta.Name] = NamespaceResource{
			PodList: podls,
			ServiceList: svcls,
		}
		log.Printf("Found %d pods and %d services in namespace %s", len(podls.Items), len(svcls.Items), namespace.ObjectMeta.Name)
	}
	return nil
}

func (c *Cache) Namespaces() *v1.NamespaceList {
	return c.NamespaceList
}

func (c *Cache) Namespace(name string) *v1.Namespace {
	for _, n := range c.NamespaceList.Items {
		if n.ObjectMeta.Name == name {
			return &n
		}
	}
	return nil
}

func (c *Cache) Pods(ns string) *v1.PodList {
	return c.NamespaceResources[ns].PodList
}

func (c *Cache) Pod(name, ns string) *v1.Pod {
	for _, p := range c.NamespaceResources[ns].PodList.Items {
		if p.ObjectMeta.Name == name {
			return &p
		}
	}
	return nil
}

func (c *Cache) Resource(name string) interface{} {
	
}

// func (c *Cache) Sync() (*Cache, error) {
	
// 	// Read kubeconfig
// 	config, err := clientcmd.BuildConfigFromFlags("", "/Users/amir/.kube/config")
// 	if err != nil {
// 		return nil, err
// 	}

// 	//Create the client
// 	clientset, err := kubernetes.NewForConfig(config)
// 	if err != nil {
// 		return nil, err
// 	}

// 	//Sync all Namespaces
// 	nsls, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	c.NamespaceList = nsls
// 	log.Printf("Found %d Namespaces", len(nsls.Items))

// 	// Sync all Pods
// 	podls, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	c.PodList = podls
// 	log.Printf("Found %d Pods", len(podls.Items))

// 	// Sync all Services
// 	svcls, err := clientset.CoreV1().Services("").List(metav1.ListOptions{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Printf("Found %d Services", len(svcls.Items))

// 	return c, nil

// }


