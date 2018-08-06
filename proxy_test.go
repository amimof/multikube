package multikube_test

import (
	"gitlab.com/amimof/multikube"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Just creates a new proxy instance
func TestProxyNewProxy(t *testing.T) {
	p := multikube.NewProxy()
	t.Logf("Config: %+v", p.Config)
}

// Test the logging middleware. Should print output to the console
func TestProxyLoggingMiddleware(t *testing.T) {
	p := multikube.NewProxy()
	m := p.Use(multikube.WithLogging)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/", nil)
	w := httptest.NewRecorder()

	m(p).ServeHTTP(w, req)

}

// Test the empty middleware. Shouldn't do anything
func TestProxyEmptyMiddleware(t *testing.T) {
	p := multikube.NewProxy()
	m := p.Use(multikube.WithEmpty)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/", nil)
	w := httptest.NewRecorder()

	m(p).ServeHTTP(w, req)
}

// Send a request through the proxy just to see that something on the other end responds
func TestProxyGetResource(t *testing.T) {
	p := multikube.NewProxy()
	m := p.Use(multikube.WithLogging)

	req := httptest.NewRequest("GET", "/api/v1/", nil)
	w := httptest.NewRecorder()

	m(p).ServeHTTP(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Status: %d", resp.StatusCode)
	t.Logf("Content-Type: %s", resp.Header.Get("Content-Type"))
	t.Logf("Body: %s", string(body))
}

// Send a request through the proxy just to see that something on the other end responds
// Currently this will wait until client closes the connections
func TestProxyWatchResource(t *testing.T) {
	p := multikube.NewProxy()
	m := p.Use(multikube.WithLogging)

	req := httptest.NewRequest("GET", "/api/v1/namespaces?watch=true", nil)
	w := httptest.NewRecorder()

	m(p).ServeHTTP(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Status: %d", resp.StatusCode)
	t.Logf("Content-Type: %s", resp.Header.Get("Content-Type"))
	t.Logf("Body: %s", string(body))
}
