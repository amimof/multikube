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

func TestGetClusterCaches(t *testing.T) {
	m := multikube.New()
	for i, cluster := range m.Clusters {
		t.Logf("Cluster: %d", i)
		t.Logf("  Name: %+v", cluster.Config.Name)
		t.Logf("    Cache ID: %s", cluster.Cache().ID)
	}
}

func TestGetClusterConfig(t *testing.T) {
	m := multikube.New()
	for i, cluster := range m.Clusters {
		t.Logf("Cluster: %d", i)
		t.Logf("  Name: %s", cluster.Config.Name)
		t.Logf("  Hostname: %s", cluster.Config.Hostname)
		t.Logf("  CA: %s", cluster.Config.CA)
		t.Logf("  Cert: %s", cluster.Config.Cert)
		t.Logf("  Key: %s", cluster.Config.Key)
	}
}

func TestSyncHTTP(t *testing.T) {
	m := multikube.New()
	for _, cluster := range m.Clusters {
		cluster.SyncHTTP()
		t.Logf("Cluster: %s", cluster.Config.Name)
		t.Logf("Cache:")
		t.Logf("  ID: %s", cluster.Cache().ID)
		t.Logf("  Store Length: %d", len(cluster.Cache().Store))
	}
}

func TestGetClusterVersion(t *testing.T) {
	m := multikube.New()
	for _, cluster := range m.Clusters {
		t.Logf("Cluster: %s", cluster.Config.Name)
		t.Logf("  Version:")
		t.Logf("    BuildDate: %s", cluster.Version().BuildDate)
		t.Logf("    Compiler: %s", cluster.Version().Compiler)
		t.Logf("    GitCommit: %s", cluster.Version().GitCommit)
		t.Logf("    GitTreeState: %s", cluster.Version().GitTreeState)
		t.Logf("    GitVersion: %s", cluster.Version().GitVersion)
		t.Logf("    GoVersion: %s", cluster.Version().GoVersion)
		t.Logf("    Major: %s", cluster.Version().Major)
		t.Logf("    Minor: %s", cluster.Version().Minor)
		t.Logf("    Platform: %s", cluster.Version().Platform)
	}
}