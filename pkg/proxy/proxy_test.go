package proxy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd/api"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	name      = "dev-cluster-1"
	defServer = "https://real-k8s-server:8443"
	defToken  = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhbWlyQG1pZGRsZXdhcmUuc2UiLCJuYW1lIjoiSm9obiBEb2UiLCJhZG1pbiI6dHJ1ZSwiaWF0IjoxNTE2MjM5MDIyfQ.nSyFTR7SZ95-pkt_PcjbmVX7rZDizLxONOnF9HWhBIe1R6ir-rrzmOaXjVxfdcVlBKEFE9bz6PJMwD8-tqsZUqlOeXSLNXXeCGhdmhluBJrJMi-Ewyzmvm7yJ2L8bVfhhBJ3z_PivSbxMKLpWz7VkbwaJrk8950QkQ5oB_CV0ysoppTybGzvU1e8tc5h5wRKimju3BA3mA5HxN8K7-2lM_JZ8cbxBToGMBMsHKSy4VXAxm-lmvSwletLXqdSlqDQZejjJYYGaPpvDih1voTJ_FJnYFzx_NWq5qN416IGJrr1RAe92B2gfRUmzftFMMw8NEYBLDNXgKx3d9OOO9xKi9DxZ9wkFrZlwNZBj-VPTgNt5zeNgME8CJqgxvCaESuDAMWkjnfdyhBYAu9uUvbRSjFowFdQFumnVlKNfAlhKOQFOZpifFIwRFYda8lzvlJv1CzHEt500HgL2qofoIOTzFQNeJ_XkOQvRBy4eBkwxKvbHlwUAObxzZrCBjaAeQRGrMU926zpujSFQ_9KzUqNsNrxJWkBybOFViQp5mMZGFIWJbdt_oiROwZLG-NDK2i932hepUfr0i52mrTX-M9vTwy4uQsiMh2eSI7Ntghw0_xgrqqp6HZON7RPdKo2ldC5_rt9TFKKmyXvhZFLgxwsm8bzvqlIbV4KwNbEZIhh-n0"
)

var backendServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `{"apiVersion":"v1","items":[],"kind":"List","metadata":{"resourceVersion":"","selfLink":""}}`)
}))

var kubeConf *api.Config = &api.Config{
	APIVersion: "v1",
	Kind:       "Config",
	Clusters: map[string]*api.Cluster{
		name: {
			Server: backendServer.URL,
		},
	},
	AuthInfos: map[string]*api.AuthInfo{
		name: {
			Token: defToken,
		},
	},
	Contexts: map[string]*api.Context{
		name: {
			Cluster:  name,
			AuthInfo: name,
		},
	},
	CurrentContext: name,
}

// Tests and empty proxy without config. Should return 502 bad gateway
func TestProxy(t *testing.T) {
	p, err := New(kubeConf)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/dev-cluster-1/api/v1/pods/default", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Got status code %d. Expected: %d", status, http.StatusOK)
	}

}

func TestProxyParseURL(t *testing.T) {
	urlString := "https://127.0.0.1:8443/api/v1/namespaces?limit=500"
	u := parseURL(urlString)
	assert.Equal(t, "127.0.0.1:8443", u.Host, "Got unexpected host in URL")
	assert.Equal(t, "/api/v1/namespaces", u.Path, "Got unexpected path in URL")
	assert.Equal(t, "limit=500", u.RawQuery, "Got unexpected query in URL")
	assert.Equal(t, "https", u.Scheme, "Got unexpected scheme in URL")
}

func TestProxy_getClusterByContextName(t *testing.T) {
	cluster := getClusterByContextName(kubeConf, "dev-cluster-1")
	assert.NotNil(t, cluster, nil)
}

func TestProxy_getAuthByContextName(t *testing.T) {
	cluster := getAuthByContextName(kubeConf, "dev-cluster-1")
	assert.NotNil(t, cluster, nil)
}

func TestProxy_SetCacheTTL(t *testing.T) {
	p, err := New(kubeConf)
	if err != nil {
		t.Fatal(err)
	}
	expected := time.Second * 12
	p.CacheTTL(expected)
	for key := range p.transports {
		assert.Equal(t, expected, p.transports[key].(*Transport).Cache.TTL, "Got unexpected cache ttl on at least 1 transport")
	}
}

func TestProxy_Use(t *testing.T) {
	p, err := New(kubeConf)
	if err != nil {
		t.Fatal(err)
	}
	p.Use(WithEmpty(), WithLogging(), WithJWT())
	assert.Equal(t, 3, len(p.middleware), "Got unexpected number of middlewares")
}

func TestProxy_Apply(t *testing.T) {

	assert := assert.New(t)
	req, err := http.NewRequest("GET", "/dev-cluster-1/api/v1/pods/default", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhbWlyQG1pZGRsZXdhcmUuc2UiLCJuYW1lIjoiSm9obiBEb2UiLCJhZG1pbiI6dHJ1ZSwiaWF0IjoxNTE2MjM5MDIyfQ.nSyFTR7SZ95-pkt_PcjbmVX7rZDizLxONOnF9HWhBIe1R6ir-rrzmOaXjVxfdcVlBKEFE9bz6PJMwD8-tqsZUqlOeXSLNXXeCGhdmhluBJrJMi-Ewyzmvm7yJ2L8bVfhhBJ3z_PivSbxMKLpWz7VkbwaJrk8950QkQ5oB_CV0ysoppTybGzvU1e8tc5h5wRKimju3BA3mA5HxN8K7-2lM_JZ8cbxBToGMBMsHKSy4VXAxm-lmvSwletLXqdSlqDQZejjJYYGaPpvDih1voTJ_FJnYFzx_NWq5qN416IGJrr1RAe92B2gfRUmzftFMMw8NEYBLDNXgKx3d9OOO9xKi9DxZ9wkFrZlwNZBj-VPTgNt5zeNgME8CJqgxvCaESuDAMWkjnfdyhBYAu9uUvbRSjFowFdQFumnVlKNfAlhKOQFOZpifFIwRFYda8lzvlJv1CzHEt500HgL2qofoIOTzFQNeJ_XkOQvRBy4eBkwxKvbHlwUAObxzZrCBjaAeQRGrMU926zpujSFQ_9KzUqNsNrxJWkBybOFViQp5mMZGFIWJbdt_oiROwZLG-NDK2i932hepUfr0i52mrTX-M9vTwy4uQsiMh2eSI7Ntghw0_xgrqqp6HZON7RPdKo2ldC5_rt9TFKKmyXvhZFLgxwsm8bzvqlIbV4KwNbEZIhh-n0")
	rr := httptest.NewRecorder()

	p, err := New(kubeConf)
	if err != nil {
		t.Fatal(err)
	}
	p.Use(WithEmpty(), WithLogging(), WithJWT())
	middleware := p.Apply(p.middleware...)

	middleware(p).ServeHTTP(rr, req)

	expected := string(`{"apiVersion":"v1","items":[],"kind":"List","metadata":{"resourceVersion":"","selfLink":""}}`)
	assert.JSONEq(expected, rr.Body.String(), "Got unexpected response body")
}
