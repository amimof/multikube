package proxy

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd/api"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestProxy(t *testing.T) {
	p := New()
	req, err := http.NewRequest("GET", "/dev-cluster-1/api/v1/pods/default", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	p.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Fatalf("Received status code '%d'. Response: '%s'", status, rr.Body.String())
	}

	// We expect 'no route' since we are not using any middleware nor is client sending any credentials
	expected := "No route: context not found\n"
	assert.Equal(t, expected, rr.Body.String(), "Got unexpected response body")
}

func TestProxyParseURL(t *testing.T) {
	urlString := "https://127.0.0.1:8443/api/v1/namespaces?limit=500"
	u := parseURL(urlString)
	assert.Equal(t, "127.0.0.1:8443", u.Host, "Got unexpected host in URL")
	assert.Equal(t, "/api/v1/namespaces", u.Path, "Got unexpected path in URL")
	assert.Equal(t, "limit=500", u.RawQuery, "Got unexpected query in URL")
	assert.Equal(t, "https", u.Scheme, "Got unexpected scheme in URL")
}

func TestProxyGetOptsFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey, "dev-cluster-1")
	ctx = context.WithValue(ctx, subjectKey, "lazy_developer")
	opts := optsFromCtx(ctx, kubeConf)
	assert.NotNil(t, opts, nil)
}
