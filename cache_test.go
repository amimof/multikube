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

func TestSyncHTTP(t *testing.T) {
	m := multikube.New()
	for _, cluster := range m.Clusters {
		cluster.Cache().SyncHTTP(&cluster)
		t.Logf("Cluster %s syncronised", cluster.Name())
	}
}