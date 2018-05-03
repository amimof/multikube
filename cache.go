package multikube

import (
	"k8s.io/api/core/v1"
	"fmt"
	//"encoding/json"
)

type Cache struct {
	Pods []*v1.Pod `json:"pods"`
}

func (c *Cache) Sync(ns string, cl *Cluster) error {
	b, err := get(fmt.Sprintf("%s/api/v1/namespaces/%s/pods/", cl.Hostname, ns))
	if err != nil {
		return err
	}
	fmt.Printf("%s", string(b))
	return nil
}