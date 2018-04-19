package multikube

import (
	"net/http"
)

type Router struct {
	Routes []*Route
	Mux *http.ServeMux
}

func NewRouter() *Router {
	return &Router{Mux: http.NewServeMux()}
}

func (r *Router) NewRoute() *Route {
	route := &Route{mux: r.Mux}
	r.Routes = append(r.Routes, route)
	return route
}

func (r *Router) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *Route {
	return r.NewRoute().Path(path).HandleFunc(f)
}

func (r *Router) Listen(a string) {
	http.ListenAndServe(a, r.Mux)
}