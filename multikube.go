package multikube

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
  "gopkg.in/yaml.v2"
	"io/ioutil"
)

func handleResponse(m *v1.Status) error {
	if m.Kind == "Status" && m.Status == "Failure" {
		return newErr(m.Message)
	}
	return nil
}

func newErrf(s string, f ...interface{}) error {
	return errors.New(fmt.Sprintf(s, f...))
}

func newErr(s string) error {
	return errors.New(s)
}

func SetupConfig(path string) (*Config, error) {

  b, err := ioutil.ReadFile(path)
  if err != nil {
    return nil, err
  }

  c := &Config{
		LogPath: "/var/log/containers",
		Clusters: []ClusterConfig{},
  }

  err = yaml.Unmarshal(b, &c)
  if err != nil {
    return nil, err
  }

  return c, nil

}

func NewGroup(name string) *Group {
	return &Group{ 
		Name: name,
		clusters: make(map[string]Cluster), 
	}
}
