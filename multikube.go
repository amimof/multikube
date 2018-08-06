package multikube

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
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

func ConfigFrom(p string) (api.Config, error) {
	b, err := ioutil.ReadFile(p)
	var c api.Config

	if err != nil {
		return c, err
	}
	fmt.Printf("%s", string(b))

	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return c, err
	}
	return c, nil
}

func SetupConfig(configPath string) (*Config, error) {

	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	c := &Config{
		LogPath:    "/var/log/multikube.log",
		APIServers: []*APIServer{},
	}

	err = json.Unmarshal(b, c)
	if err != nil {
		return nil, err
	}

	return c, nil

}

// NewCache return a new empty cache
func NewCache() *Cache {
	return &Cache{
		ID:    uuid.New(),
		Store: make(map[string]Item),
	}
}
