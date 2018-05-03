package multikube_test

import (
	"testing"
	"github.com/amimof/multikube"
)

func TestGetConfig(t *testing.T) {
	c := multikube.SetupConfig()
	t.Logf("LogPath: %s", c.LogPath)
	for i, cluster := range c.Clusters {
		t.Logf("Cluster: %d", i)
		t.Logf("Name: %s", cluster.Name)
		t.Logf("ID: %d", cluster.ID)
		t.Logf("UUID: %s", cluster.UUID)
		t.Logf("Hostname: %s", cluster.Hostname)
		t.Logf("Credential: %+v", cluster.Credential)
	}
}