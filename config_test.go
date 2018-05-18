package multikube_test

import (
	"testing"
	"gitlab.com/amimof/multikube"
)

func TestGetConfig(t *testing.T) {
	c := multikube.SetupConfig()
	t.Logf("LogPath: %s", c.LogPath)
	for i, cluster := range c.Clusters {
		t.Logf("Cluster: %d", i)
		t.Logf("  Name: %s", cluster.Name)
		t.Logf("  Hostname: %s", cluster.Hostname)
		t.Logf("  CA %s", cluster.CA)
		t.Logf("  Cert %s", cluster.Cert)
		t.Logf("  Key %s", cluster.Key)
	}
}