package multikube

import (
	"strings"
	"io/ioutil"
	"net/http"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/gorilla/mux"
)

type Multikube struct {
	Version string
	Config *Config
	router *mux.Router 
}

type APIErrorResponse struct {
	Code int
	Err error
}

func (a APIErrorResponse) Error() string {
	return a.Err.Error()
}

func (a APIErrorResponse) Status() int {
	return a.Code
}

func handleErr(w http.ResponseWriter, err error) {
	if err != nil {
		switch e := err.(type) {
		case APIErrorResponse:
			http.Error(w, e.Error(), e.Status())
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
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

func getSSL(url string, cl *Cluster) ([]byte, error) {

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(cl.Cert, cl.Key)
	if err != nil {
		return nil, err
	}

	// Load CA certificate
	caCert, err := ioutil.ReadFile(cl.CA)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs: caCertPool,
	}

	tlsConfig.BuildNameToCertificate()
	tr := &http.Transport{ TLSClientConfig: tlsConfig }
	client := &http.Client{ Transport: tr }

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func get(url string) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: tr,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func put(url string, data []byte) ([]byte, error) {

	body := strings.NewReader(string(data))

	req, err := http.NewRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func post(url string, data []byte) ([]byte, error) {

	body := strings.NewReader(string(data))

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func delete(url string) ([]byte, error) {

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return result, nil

}

func (m *Multikube) SetupRoutes() *Multikube {

	// Cluster
	m.router.HandleFunc("/clusters", m.GetClustersHandler).Methods("GET")
	m.router.HandleFunc("/clusters", m.CreateClusterHandler).Methods("POST")
	m.router.HandleFunc("/clusters/{name}", m.GetClusterHandler).Methods("GET")
	m.router.HandleFunc("/clusters/{name}", m.DeleteClusterHandler).Methods("DELETE")
	m.router.HandleFunc("/clusters/{name}", m.UpdateClusterHandler).Methods("PUT")
	
	// Namespace
	m.router.HandleFunc("/clusters/{name}/namespaces", m.GetClusterNamespacesHandler).Methods("GET")
	m.router.HandleFunc("/clusters/{name}/namespaces/{ns}", m.GetClusterNamespaceHandler).Methods("GET")
	
	// Pods
	m.router.HandleFunc("/clusters/{name}/namespaces/{ns}/pods", m.GetClusterPodsHandler).Methods("GET")
	m.router.HandleFunc("/clusters/{name}/namespaces/{ns}/pods/{pod}", m.GetClusterPodHandler).Methods("GET")

	return m
}

func (m *Multikube) ListenAndServe(addr string) {
	http.ListenAndServe(":8081", m.router)
}

func New() *Multikube {
	return &Multikube{
		Version: "1.0.0",
		Config: SetupConfig(),
		router: mux.NewRouter(),
	}
}