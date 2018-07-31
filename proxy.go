package multikube

import (
	//"io"
	"log"
	"net"
	"context"
	"net/http"
	"crypto/tls"
	"net/http/httputil"
	"github.com/spf13/pflag"
	//"io/ioutil"
	//"bufio"
)

var (
	configPath string
)

type Proxy struct {
	Config  *Config
}

func init() {
	pflag.StringVar(&configPath, "config", "/etc/multikube/multikube.yaml", "Path to the multikube configuration")
}

// NewProxy crerates a new Proxy and initialises router and configuration
func NewProxy() *Proxy {

	// Read config from disk
	c, err := SetupConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Define API
	p := &Proxy{c}

	// Apply middleware
	p.Use(
		WithEmpty,
		WithLogging,
	)

	return p
}

// Use chains all middlewares and applies a context to the request flow
func (p *Proxy) Use(mw ...MiddlewareFunc) MiddlewareFunc {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](last)
			}
			ctx := context.WithValue(r.Context(), "config", p.Config)
			last(w, r.WithContext(ctx))
		}
	}
}


// Works except Watch
//
// proxy routes the request to an apiserver. It determines resolves an apiserver using
// data in the request itsel such as certificate data, authorization bearer tokens, http headers etc.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Header.Get("Upgrade") != "" {
		p.tunnel(w, r)
		return
	}
	
	// Build the request and execute the call to the backend apiserver
	// Config is hardcoded. We need a way of determining an apiserver
	// based on cert data, token or headers.
	// Might need a middleware that propagates context before calling proxy() function.
	req := 
		NewRequest(p.Config.APIServers[1]).
		Method(r.Method).
		Body(r.Body).
		Path(r.URL.Path).
		Query(r.URL.RawQuery).
		Headers(r.Header)

	// Execute!
	res, err := req.Do()
	defer res.Body.Close()

	// Catch any unexpected errors
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", req.Response().Header.Get("Content-Type"))
	w.WriteHeader(res.StatusCode)

	buf := make([]byte, 4096)
	for {
		n, err := res.Body.Read(buf)
		if n == 0 && err != nil {
			break
		}
		b := buf[:n]
		w.Write(b)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}		
	
}

// tunnel hijacks the client request, creates a pipe between client and backend server
// and starts streaming data between the two connections.
func (p *Proxy) tunnel(w http.ResponseWriter, r *http.Request) {

	req := NewRequest(p.Config.APIServers[1])
	tlsConfig := req.TLSConfig()

	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	dst_conn, err := tls.Dial("tcp", "192.168.99.100:8443", tlsConfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	dst_conn.Write(dump)
	
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	src_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	go transfer(dst_conn, src_conn)
	go transfer(src_conn, dst_conn)

}

// transfer reads the data from src into a buffer before it writes it into dst
func transfer(src, dst net.Conn) {
	buff := make([]byte, 65535)
	defer src.Close()
	defer dst.Close()

	for {
		n, err := src.Read(buff)
		if err != nil {
			break
		}
		b := buff[:n]
		_, err = dst.Write(b)
		if err != nil {
			break
		}	
	}

	log.Printf("Transfered src: %s dst: %s bytes: %d", src.LocalAddr().String(), dst.RemoteAddr().String(), len(buff))
}