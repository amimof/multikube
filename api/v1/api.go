package v1

import (
	"log"
	"net/http"
)

type API struct {
	version string
}

func NewAPI() *API {
	return &API{ version: "v1" }
}

func (a *API) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log.Printf("Serving API %s", a.version)
}