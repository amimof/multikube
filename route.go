package multikube

import (
	"net/http"
)

type Route struct {
	path string
	mux *http.ServeMux
}

func (r *Route) Path(p string) *Route {
	r.path = p
	return r
}

func (r *Route) HandleFunc(f func(http.ResponseWriter, *http.Request)) *Route {
	r.mux.HandleFunc(r.path, f)
	return r
}