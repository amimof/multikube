package v1

import (
	"context"
	"github.com/spf13/pflag"
	"gitlab.com/amimof/multikube"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	//"time"
	"crypto/tls"
	"io"
	//"io/ioutil"
	//"bufio"
	//"bufio"
	//"k8s.io/client-go/rest"
	//"k8s.io/client-go/tools/clientcmd"
	//"k8s.io/client-go/kubernetes"
	//"golang.org/x/net/http2"
	//"io"
)

var (
	configPath string
	kubeconfigPath string
)

type API struct {
	Path    string
	Version string
	//Router  *http.ServeMux
	Config  *multikube.Config
}

func init() {
	pflag.StringVar(&configPath, "config", "/etc/multikube/multikube.yaml", "Path to the multikube configuration")
	pflag.StringVar(&kubeconfigPath, "kubeconfig", "~/.kube/config", "Path a kubeconfig file")
}

// MewAPI crerates a new API and initialises router and configuration
func NewAPI() *API {

	// Read config from disk
	c, err := multikube.SetupConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Define API
	api := &API{
		Path:    "/api/v1",
		Version: "v1",
		//Router:  http.NewServeMux(),
		Config:  c,
	}

	// Setup middlewares in order
	// mw := api.Use(
	// 	multikube.WithEmpty,
	// 	multikube.WithLogging,
	// )

	// Handle all requests here through the proxy
	//api.Router.HandleFunc("/", mw(api.routeHTTP))

	return api
}

// Use chains all middlewares and applies a context to the request flow
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

// Works except Watch
//
// proxy routes the request to an apiserver. It determines resolves an apiserver using
// data in the request itsel such as certificate data, authorization bearer tokens, http headers etc.
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("%s %s %s %s %s", r.Method, r.URL.Path, r.URL.RawQuery, r.RemoteAddr, r.Proto)

	if pusher, ok := w.(http.Pusher); ok {
		// Push is supported.
		if err := pusher.Push("/app.js", nil); err != nil {
				log.Printf("Failed to push: %s", err)
		}
	}

	log.Printf("--- CLIENT REQUEST START ---")
	for k, _ := range r.Header {
		log.Printf("%s: %s", k, r.Header.Get(k))
	}
	log.Printf("Method: %s", r.Method)
	log.Printf("--- CLIENT REQUEST END ---")

	if r.Header.Get("Upgrade") != "" {
		a.tunnel(w, r)
		return
	}

	// This part is hardcoded. We need a way of determining an apiserver
	// based on cert data, token or headers.
	// Might need a middleware that propagates context before calling proxy() function.
	//config := r.Context().Value("config").(*multikube.Config)

	// Build the request and execute the call to the backend apiserver
	req := multikube.
		NewRequest(a.Config.APIServers[1]).
		Method(r.Method).
		Body(r.Body).
		Path(r.URL.Path).
		Query(r.URL.RawQuery).
		Headers(r.Header)

	// Execute!
	res, err := req.Doo()
	//defer res.Body.Close()

	// Catch any unexpected errors
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error: %s", err)
		return
	}

	w.Header().Set("Content-Type", req.Response().Header.Get("Content-Type"))
	w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)
	
}

func (a *API) tunnel(w http.ResponseWriter, r *http.Request) {

	//config := r.Context().Value("config").(*multikube.Config)
	req := multikube.NewRequest(a.Config.APIServers[1])
	tlsConfig := req.TLSConfig()

	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Printf("Unable to dump: %s", err)
		return
	}

	// log.Printf("--- TUNNEL START ---")
	// for k, _ := range r.Header {
	// 	log.Printf("%s: %s", k, r.Header.Get(k))
	// }
	// log.Printf("Method: %s", r.Method)
	// log.Printf("--- TUNNEL END ---")

	dst_conn, err := tls.Dial("tcp", "192.168.99.100:8443", tlsConfig)
	if err != nil {
		log.Printf("Unable to dial downstream")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	dst_conn.Write(dump)
	
	//w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	src_conn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("Unable to hijack connection")
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	go transfer(dst_conn, src_conn, "dest -> source")
	go transfer(src_conn, dst_conn, "source -> dest")

}

func transferBody(destination http.ResponseWriter, source io.ReadCloser) {
	defer source.Close()
	n, err := io.Copy(destination, source)
	if err != nil {
		log.Printf("Error: %s", err)
	}

	log.Printf("Copied %d bytes", n)
}

func transfer(src, dst net.Conn, name string) {
	buff := make([]byte, 65535)
	defer src.Close()
	defer dst.Close()
	for {

		log.Printf("%s Buffer: %s", name, string(buff))
		n, err := src.Read(buff)
		if err == io.EOF {
			continue
		}
		if err != nil {
			log.Printf("Error Read: %s: %s", name, err)
		}
		log.Printf("%s: Read %d bytes", name, n)

		b := buff[:n]
		w, err := dst.Write(b)
		if err != nil {
			log.Printf("Error Write: %s: %s", name, err)
		}

		log.Printf("%s: Wrote %d bytes", name, w)

	}
}