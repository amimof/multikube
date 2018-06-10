package v1

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
)

type API struct {
	Router *mux.Router
	Path string
	Version string
}

func NewAPI() *API {
	api := &API{
		Router: mux.NewRouter(),
		Path: "/api/v1",
		Version: "v1",
	}
	s := api.Router.PathPrefix(api.Path).Subrouter()
	
	// Register our routes
	s.HandleFunc("/clusters", List).Methods("GET")
	s.HandleFunc("/clusters/{name}", Get).Methods("GET")
	s.HandleFunc("/clusters/", Create).Methods("POST")
	s.HandleFunc("/clusters/{name}", Update).Methods("PUT")

	return api
}

func (a *API) Handler() http.Handler {
	return a.Router
}

func (a *API) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log.Println("Serving API v1")
}