package multikube

import (
	//"net/http"
	// "strings"
	// "io/ioutil"
	// "crypto/tls"
	// "crypto/x509"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Multikube struct {
	Version string
	Config *Config
	Clusters []Cluster
}

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

// func getSSL(url string, config *ClusterConfig) ([]byte, error) {

// 	// Load client certificate
// 	cert, err := tls.LoadX509KeyPair(config.Cert, config.Key)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Load CA certificate
// 	caCert, err := ioutil.ReadFile(config.CA)
// 	if err != nil {
// 		return nil, err
// 	}

// 	caCertPool := x509.NewCertPool()
// 	caCertPool.AppendCertsFromPEM(caCert)

// 	tlsConfig := &tls.Config{
// 		Certificates: []tls.Certificate{cert},
// 		RootCAs: caCertPool,
// 	}

// 	tlsConfig.BuildNameToCertificate()
// 	tr := &http.Transport{ TLSClientConfig: tlsConfig }
// 	client := &http.Client{ Transport: tr }

// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	res, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer res.Body.Close()

// 	body, err := ioutil.ReadAll(res.Body)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return body, nil
// }

// func get(url string) ([]byte, error) {

// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	tr := &http.Transport{
// 		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
// 	}
// 	client := http.Client{
// 		Transport: tr,
// 	}
// 	res, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer res.Body.Close()

// 	body, err := ioutil.ReadAll(res.Body)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return body, nil
// }

// func put(url string, data []byte) ([]byte, error) {

// 	body := strings.NewReader(string(data))

// 	req, err := http.NewRequest("PUT", url, body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")

// 	client := http.Client{}
// 	res, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer res.Body.Close()

// 	result, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil

// }

// func post(url string, data []byte) ([]byte, error) {

// 	body := strings.NewReader(string(data))

// 	req, err := http.NewRequest("POST", url, body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")

// 	client := http.Client{}
// 	res, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer res.Body.Close()

// 	result, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil

// }

// func delete(url string) ([]byte, error) {

// 	req, err := http.NewRequest("DELETE", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")

// 	client := http.Client{}
// 	res, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer res.Body.Close()

// 	result, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return result, nil

// }

func (m *Multikube) GetClusters() ([]Cluster, error) {
	return m.Clusters, nil
}

func (m *Multikube) GetCluster(name string) (*Cluster, error) {
	for _, cluster := range m.Clusters {
		if cluster.Config.Name == name {
			return &cluster, nil
		}
	}
	return nil, newErrf("Cluster %s does not exist", name)
}

func New() *Multikube {
	var clusters []Cluster
	c := SetupConfig()
	
	for _, config := range c.Clusters {
		clusters = append(clusters, Cluster{Config: &config})
	}

	return &Multikube{
		Version: "1.0.0",
		Config: c,
		Clusters: clusters,
	}
}

func NewForConfig(c *Config) *Multikube {
	clusters := make([]Cluster, len(c.Clusters))
	for i, config := range c.Clusters {
		clusters[i].Config = &config
	}
	return &Multikube{
		Version: "1.0.0",
		Config: c,
		Clusters: clusters,
	}
}