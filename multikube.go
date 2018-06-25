package multikube

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"net/http"
	"encoding/json"
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

func NewGroup(name string) *Group {
	return &Group{ 
		Name: name,
		clusters: []Cluster{}, 
	}
}

// NewContext returns a Context instance.
func NewContext(w http.ResponseWriter, r *http.Request) Context {
	return Context{
		Request:  r,
		Response: &w,
	}
}

func SetupConfig(configPath string) (*Config, error) {

  b, err := ioutil.ReadFile(configPath)
  if err != nil {
    return nil, err
  }

  c := &Config{
		LogPath: "/var/log/multikube.log",
		APIServers: []APIServer{},
  }

  err = json.Unmarshal(b, &c)
  if err != nil {
    return nil, err
  }

  return c, nil

}