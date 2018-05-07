package multikube

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/tools/cache"
	//"k8s.io/apimachinery/pkg/watch"
	//"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/fields"
	"log"
	"encoding/json"
)

type Cache struct {
	NamespaceList *v1.NamespaceList `json:"namespaces"`
	PodList *v1.PodList `json:"pods,omitempty"`
	ServiceList *v1.ServiceList `json:"services,omitempty"`
	client *Request
}


func (c *Cache) Sync() (*Cache, error) {
	
	// Read kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/amir/.kube/config")
	if err != nil {
		return nil, err
	}

	//Create the client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	//Sync all Namespaces
	nsls, err := clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	c.NamespaceList = nsls
	log.Printf("Found %d Namespaces", len(nsls.Items))

	// Sync all Pods
	podls, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	c.PodList = podls
	log.Printf("Found %d Pods", len(podls.Items))

	// Sync all Services
	svcls, err := clientset.CoreV1().Services("").List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	log.Printf("Found %d Services", len(svcls.Items))

	return c, nil

}


func (c *Cache) SyncHTTP(cl *Cluster) error {

	r, err := NewRequest(cl).Get().Resource("namespaces")
	if err != nil {
		return err
	}

	r, err = r.Do()
	if err != nil {
		return err
	}

	resp := &v1.NamespaceList{}
	err = json.Unmarshal(r.Data(), &resp)
	if err != nil {
		return err
	}

	c.NamespaceList = resp

	return nil

	// b, err := getSSL(fmt.Sprintf("%s/api/v1/namespaces", cl.Hostname), cl)
	// if err != nil {
	// 	return err
	// }
	// resp := &metav1.Status{}
	// err = json.Unmarshal(b, &resp)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(string(b))
	// fmt.Printf("%+v", resp)
	// err = handleResponse(resp)
	// if err != nil {
	// 	return err
	// }
	// return nil
}