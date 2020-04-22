package proxy

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareWithHeader(t *testing.T) {
	assert := assert.New(t)

	p := New().Use(WithHeader())
	p.KubeConfig = kubeConf

	req, err := http.NewRequest("GET", "/api/v1/pods/default", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	req.Header.Set("Multikube-Context", "dev-cluster-1")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhbWlyQG1pZGRsZXdhcmUuc2UiLCJuYW1lIjoiSm9obiBEb2UiLCJhZG1pbiI6dHJ1ZSwiaWF0IjoxNTE2MjM5MDIyfQ.nSyFTR7SZ95-pkt_PcjbmVX7rZDizLxONOnF9HWhBIe1R6ir-rrzmOaXjVxfdcVlBKEFE9bz6PJMwD8-tqsZUqlOeXSLNXXeCGhdmhluBJrJMi-Ewyzmvm7yJ2L8bVfhhBJ3z_PivSbxMKLpWz7VkbwaJrk8950QkQ5oB_CV0ysoppTybGzvU1e8tc5h5wRKimju3BA3mA5HxN8K7-2lM_JZ8cbxBToGMBMsHKSy4VXAxm-lmvSwletLXqdSlqDQZejjJYYGaPpvDih1voTJ_FJnYFzx_NWq5qN416IGJrr1RAe92B2gfRUmzftFMMw8NEYBLDNXgKx3d9OOO9xKi9DxZ9wkFrZlwNZBj-VPTgNt5zeNgME8CJqgxvCaESuDAMWkjnfdyhBYAu9uUvbRSjFowFdQFumnVlKNfAlhKOQFOZpifFIwRFYda8lzvlJv1CzHEt500HgL2qofoIOTzFQNeJ_XkOQvRBy4eBkwxKvbHlwUAObxzZrCBjaAeQRGrMU926zpujSFQ_9KzUqNsNrxJWkBybOFViQp5mMZGFIWJbdt_oiROwZLG-NDK2i932hepUfr0i52mrTX-M9vTwy4uQsiMh2eSI7Ntghw0_xgrqqp6HZON7RPdKo2ldC5_rt9TFKKmyXvhZFLgxwsm8bzvqlIbV4KwNbEZIhh-n0")

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	p.Chain().ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Received status code '%d'. Response: '%s'", status, rr.Body.String())
	}

	// Check the response body is what we expect.
	expected := string(`{"apiVersion":"v1","items":[],"kind":"List","metadata":{"resourceVersion":"","selfLink":""}}`)
	assert.JSONEq(expected, rr.Body.String(), "Got unexpected response body")

}
