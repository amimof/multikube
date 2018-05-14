package multikube

import (
  "gopkg.in/yaml.v2"
  "os"
  "log"
  "io/ioutil"
  //"path"
  //"net/url"
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

func SetupConfig() *Config {

  var configPath string = "/etc/multikube/multikube.yaml"
  if !exists(configPath) {
      file, err := os.Create(configPath)
      if err != nil {
        log.Fatal(err)
      }
      defer file.Close()
  }

  b, err := ioutil.ReadFile(configPath)
  if err != nil {
    log.Fatal(err)
  }

  c := &Config{
		LogPath: "/var/log/containers",
		Clusters: []ClusterConfig{},
  }

  err = yaml.Unmarshal(b, &c)
  if err != nil {
    log.Fatal(err)
  }

  return c

}

// func (c *Config) GetApiUrl(str ...string) string {
// 	u, err := url.Parse(c.KubernetesApi)
// 	if err != nil {
// 		return ""
// 	}
// 	u.Path = path.Join(u.Path, "/api/v1/")
// 	for _, p := range str {
// 		u.Path = path.Join(u.Path, p)
// 	}
// 	return u.String()
// }
