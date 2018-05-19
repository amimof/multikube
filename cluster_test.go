package multikube_test

import (
	"testing"
)

func TestNewCluster(t *testing.T) {
	newcluster := group.Clusters()[0]
	newcluster.Config.Name = "minikube_new"
	group = group.AddCluster(newcluster.Config)
	for i, cluster := range group.Clusters() {
		t.Logf("Cluster %d: %s", i, cluster.Config.Name)
	}
}

func TestGetClusters(t *testing.T) {
	for i, cluster := range group.Clusters() {
		t.Logf("Cluster %d: %s", i, cluster.Config.Name)
	}
}

func TestGetClusterCaches(t *testing.T) {
	for i, cluster := range group.Clusters() {
		t.Logf("Cluster: %d", i)
		t.Logf("  Name: %+v", cluster.Config.Name)
		t.Logf("    Cache ID: %s", cluster.Cache().ID)
	}
}

func TestGetClusterConfig(t *testing.T) {
	for i, cluster := range group.Clusters() {
		t.Logf("Cluster: %d", i)
		t.Logf("  Name: %s", cluster.Config.Name)
		t.Logf("  Hostname: %s", cluster.Config.Hostname)
		t.Logf("  CA: %s", cluster.Config.CA)
		t.Logf("  Cert: %s", cluster.Config.Cert)
		t.Logf("  Key: %s", cluster.Config.Key)
	}
}

func TestSyncHTTP(t *testing.T) {
	for _, cluster := range group.Clusters() {
		cache, err := cluster.SyncHTTP()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Cluster: %s", cluster.Config.Name)
		t.Logf("Cache:")
		t.Logf("  ID: %s", cache.ID)
		t.Logf("  Store Length: %d", len(cache.Store))
		t.Logf("  Store Size: %d", cache.Size())
		t.Logf("  Keys:")
		for _, key := range cache.ListKeys() {
			t.Logf("    %s: %d bytes", key, cache.Get(key).Bytes())
		}
	}
}

func TestGetClusterVersion(t *testing.T) {
	for _, cluster := range group.Clusters() {
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