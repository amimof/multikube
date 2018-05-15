package multikube_test

import (
	"testing"
	"github.com/amimof/multikube"
)

func TestGetCaches(t *testing.T) {
	m := multikube.New()
	for _, cluster := range m.Clusters {
		t.Logf("Cluster: %s, Cache: %s", cluster.Name(), cluster.Cache().ID)
	}
}

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