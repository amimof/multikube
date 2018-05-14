package multikube

import (
	//"k8s.io/api/core/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/client-go/tools/clientcmd"
	// "k8s.io/client-go/kubernetes"
	//"k8s.io/client-go/tools/cache"
	//"k8s.io/apimachinery/pkg/watch"
	//"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/fields"
	//"log"
	"github.com/google/uuid"
)

// Root cache object
type Cache struct {
	ID uuid.UUID
	client *Request
}

func (c *Cache) SyncHTTP(cl *Cluster) *Cache {
	return c
}
