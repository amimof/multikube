package multikube

import (
	"net/http"
)

type Credential struct {
	Name string `json:"name,omitempty`
	Token string `json:"token,omitempty`
	ClientCertificate string `json:"clientcertificate,omitempty"`
	ClientKey string `json:"clientkey,omitempty"`
	ClientCertificateData string `json:"clientcertificatedata,omitempty"`
}

func GetCredentials(res http.ResponseWriter, req *http.Request) {

}

func GetCredential(res http.ResponseWriter, req *http.Request) {

}

func CreateCredential(res http.ResponseWriter, req *http.Request) {

}

func UpdateCredential(res http.ResponseWriter, req *http.Request) {

}

func DeleteCredential(res http.ResponseWriter, req *http.Request) {

}