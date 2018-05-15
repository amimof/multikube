package multikube_test

import (
	"testing"
	"github.com/amimof/multikube"
)

func TestGetItem(t *testing.T) {
	m := multikube.New()
	cluster := m.Clusters[0]
	item := cluster.Cache().Get("namespaces")
	t.Logf("Item: %+v", item)
}

func TestSetItem(t *testing.T) {
	m := multikube.New()
	cluster := m.Clusters[0]
	item := cluster.Cache().Set("namespaces", "Hello World")
	t.Logf("Item: %+v", item)
}

func TestDeleteItem(t *testing.T) {
	m := multikube.New()
	cluster := m.Clusters[0]
	item := cluster.Cache().Set("namespaces", "hello world")
	t.Logf("Item: %+v", item)
	cluster.Cache().Delete(item.Key)
	item = cluster.Cache().Get("namespaces")
	t.Logf("Item: %+v", item)
}

func TestListKeys(t *testing.T) {
	m := multikube.New()
	cluster := m.Clusters[0]
	cluster.Cache().Set("/namespaces/", []byte{'a'})
	cluster.Cache().Set("/namespaces/pods", []byte{'b'})
	cluster.Cache().Set("/namespaces/pods/pod-1", []byte{'c'})
	for _, key := range cluster.Cache().ListKeys() {
		t.Logf("%s", key)
	}
}