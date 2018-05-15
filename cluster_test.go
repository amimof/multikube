package multikube_test

import (
	"testing"
	"github.com/amimof/multikube"
)

func TestGetClusters(t *testing.T) {
	m := multikube.New()
	for i, cluster := range m.Clusters {
		t.Logf("Cluster: %d", i)
		t.Logf("  Cache: %+v", cluster.Cache())
		t.Logf("  Config: %+v", cluster.Config)
	}
}

func TestSyncHTTP(t *testing.T) {
	m := multikube.New()
	for _, cluster := range m.Clusters {
		cluster.SyncHTTP()
		t.Logf("Cluster %s syncronised", cluster.Name())
	}
}