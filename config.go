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
	Clusters []*Cluster `yaml:"clusters"`
  //PosPath string `yaml:"posPath"`
  //MaxWorkers int `yaml:"waxWorkers"`
  //KubernetesApi string `yaml:"kubernetesApi"`
  //CertFile string `yaml:certFile`
  //CaFile string `yaml:caFile`
  //KeyFile string `yaml:keyFile`
  //RequireMeta bool `yaml:requireMetadata`
  //SkipVerifySsl bool `yaml:skipVerifySsl`
  //DbHostSuffix string `yaml:dbHostSuffix`
  //DbHostPrefix string `yaml:dbHostPrefix`
  //MetadataMaxTries int `yaml:metadataMaxTries`
  //NumRecordsPerChunk int `yaml:numRecordsPerChunk`
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
		Clusters: []*Cluster{},
    //PosPath: "/tmp/kubelogs.log.pos",
    //MaxWorkers: 5,
    //KubernetesApi: "https://localhost",
    //RequireMeta: true,
    //SkipVerifySsl: false,
    //CertFile: "/etc/kubelogs/cert.crt",
    //CaFile: "/etc/kubelogs/ca.crt",
    //KeyFile: "/etc/kubelogs/cert.key",
    //DbHostPrefix: "logsdb.",
    //DbHostSuffix: ".svc.cluster.local",
    //MetadataMaxTries: 5,
    //NumRecordsPerChunk: 20,
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
