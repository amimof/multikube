package v1

import (
	"log"
	"net/http"
	"github.com/spf13/pflag"
	"gitlab.com/amimof/multikube"
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
	mw := api.Use(multikube.WithContext, multikube.WithLogging)
	
	// Handle all requests here through the proxy
	api.Router.HandleFunc("/", mw(api.proxy))

	return api
}

func (a *API) Use(mw ...multikube.MiddlewareFunc) multikube.MiddlewareFunc {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](last)
			}
			last(w, r)
		}
	}
}

func (a *API) proxy(w http.ResponseWriter, r *http.Request) {
	context := &multikube.Options{
		Hostname: "https://192.168.99.100:8443",
		CA: "/Users/amir/.minikube/ca.pem",
		Cert: "/Users/amir/.minikube/cert.pem",
		Key: "/Users/amir/.minikube/key.pem",
	}
	_, err := multikube.NewRequest(context).Get().Path(r.URL.Path).Do()
	if err != nil {
		log.Printf("Error: %s", err)
	}
}

