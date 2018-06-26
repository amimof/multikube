package v1

import (
	"log"
	"net/http"
	"github.com/spf13/pflag"
	"gitlab.com/amimof/multikube"
	"context"
)

var (
	configPath string
)

type API struct {
	Path string
	Version string
	// Proxy *httputil.ReverseProxy
	Router *http.ServeMux
	Config *multikube.Config
}

func init() {
	pflag.StringVar(&configPath, "config", "/etc/multikube/multikube.json", "Path to the multikube configuration")
}

func NewAPI() *API {

	c, err := multikube.SetupConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	
	api := &API{
		Path: "/api/v1",
		Version: "v1",
		Router: http.NewServeMux(),
		Config: c,
	}
	
	// Setup middlewares in order
	mw := api.Use(
		multikube.WithEmpty, 
		multikube.WithLogging,
	)
	
	// Handle all requests here through the proxy
	api.Router.HandleFunc("/", mw(api.proxy))

	return api
}

// Use chains all middlewars and applies a context to the request flow
func (a *API) Use(mw ...multikube.MiddlewareFunc) multikube.MiddlewareFunc {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](last)
			}
			ctx := context.WithValue(r.Context(), "config", a.Config)
			last(w, r.WithContext(ctx))
		}
	}
}

// proxy routes the request to an apiserver. It determines resolves an apiserver using
// data in the request itsel such as certificate data, authorization bearer tokens, http headers etc.
func (a *API) proxy(w http.ResponseWriter, r *http.Request) {

	// This part is hardcoded. We need a way of determining an apiserver
	// based on cert data, token or headers. 
	// Might need a middleware that propagates context before calling this function.
	options := &multikube.Options{
		Hostname: "https://192.168.99.100:8443",
		CA: "/Users/amir/.minikube/ca.crt",
		Cert: "/Users/amir/.minikube/client.crt",
		Key: "/Users/amir/.minikube/client.key",
	}

	// Build the request and execute the call to the backend apiserver
	req, err := multikube.NewRequest(options).Method(r.Method).Body(r.Body).Path(r.URL.Path).SetHeader("Content-Type", "application/json").Do()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error: %s", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(req.Data())
}