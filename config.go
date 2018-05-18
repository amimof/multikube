package multikube

import (
  "os"
)

type Config struct {
	LogPath string `yaml:"logPath"`
	Clusters []ClusterConfig `yaml:"clusters"`
}

type ClusterConfig struct {
	Name string `json:"name,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	CA string `json:"ca,omitempty"`
	Cert string `json:"cert,omitempty"`
	Key string `json:"key,omitempty"`
}

// Returns true if path exists
func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}