package multikube

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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

func NewGroup(name string) *Group {
	return &Group{ 
		Name: name,
		clusters: []Cluster{}, 
	}
}
