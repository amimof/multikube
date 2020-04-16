package proxy

import (
	"github.com/SermoDigital/jose/crypto"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var validRS256pubkey = []byte(`-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAwTlp6YkFJlrhSJ7ukHJv
wNe4+uUdTsizGK1u8Dh58EKpJkR9GrGLaB3LABC+CJpheBc5JU+Hd4UglEWIEFHK
LSAEnsGXhl7bnuNeXxzBBM3LHTRfxQe/2rj69xuW7vbZ8pbhZ1+FVsxkznm28u6F
zddAq2gLHCa0+Tc/IqqVTHx102fqzmOFMwLRHzTxoXaAx1uoRkngRK+8N3btWpQd
hz1vHNPa1+6shuhPILpgGhcyGVsiLO3v4ZUdVZTw71295wTtPCLxoM/9F3o4VaRg
dcrn9jTEUH/2uGgLNMlfpkbZaPk7p1GGaGjgaTZmFs25DurJjOADuNhiT+LXDLgo
O6PIgIBlU6CUwcs7x9TZ1N7bqpWUOOVvIyZ65UNFIbExlAJPNENOer7voG+FJJ8W
LYqmW/xGWG8sDsFZjHSpGaNq1do8eWa6y1X3eZfK7hmYWWxHDG4+0Rfcuf5lIUGq
ChD6j7cVfJf4qRJNxHSccemM8H97MYKuHPTQM0NZruUPDDpKbwelVzglBT5PgNuv
adPggesVbCunCIMggg2Wq47i551A+7Rb3Dki7FzjrHKiuv5CL+oGgxSN8jmCZXfc
jIvzLIFNEY7lxHTyccY/YNgjNEMyUeu+9qI1sy5sAoco9HSncQIrVSD+VXtIhB4n
rL/XzEalgKsZo5Z0rBmo7ycCAwEAAQ==
-----END PUBLIC KEY-----`)

var invalidRS256pubkey = []byte(`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA21sntvLQsX8p+E7uwJLM
MCyJaMn21rR8Bb2LttyZMt94xV7YKBuWKO1Y+9Qzy316qKdGYrMHSTKreear+g4B
QBYvkom4vPReMZH+BW6sYavTyNqt0Akm+PmH/E8qRDIpvkXbA4goy/tM9Bychaxm
JAKtPIVoXvdpbfmYML5XX5pC8zJuZTCMfu36ncgV+Bxyzf859uIe4oqxVxXMsbKk
wK81kodtG5WeMYcO8xtHMfwtI97IMOlvN/3VUZMWc/wpiOE3CutkaQc/wdRQtQks
fFGXHj1zUev2eB4zO+m7ks4zMCL58jIE1s1LlpE/lcEscIc8jPV6WGHuuUCnL7lB
8wIDAQAB
-----END PUBLIC KEY-----`)

func TestMiddlewareWithRS256Validation(t *testing.T) {
	assert := assert.New(t)
	req, err := http.NewRequest("GET", "/dev-cluster-1/api/v1/pods/default", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	req.Header.Set("Multikube-Context", "dev-cluster-1")
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhbWlyQG1pZGRsZXdhcmUuc2UiLCJuYW1lIjoiSm9obiBEb2UiLCJhZG1pbiI6dHJ1ZSwiaWF0IjoxNTE2MjM5MDIyfQ.nSyFTR7SZ95-pkt_PcjbmVX7rZDizLxONOnF9HWhBIe1R6ir-rrzmOaXjVxfdcVlBKEFE9bz6PJMwD8-tqsZUqlOeXSLNXXeCGhdmhluBJrJMi-Ewyzmvm7yJ2L8bVfhhBJ3z_PivSbxMKLpWz7VkbwaJrk8950QkQ5oB_CV0ysoppTybGzvU1e8tc5h5wRKimju3BA3mA5HxN8K7-2lM_JZ8cbxBToGMBMsHKSy4VXAxm-lmvSwletLXqdSlqDQZejjJYYGaPpvDih1voTJ_FJnYFzx_NWq5qN416IGJrr1RAe92B2gfRUmzftFMMw8NEYBLDNXgKx3d9OOO9xKi9DxZ9wkFrZlwNZBj-VPTgNt5zeNgME8CJqgxvCaESuDAMWkjnfdyhBYAu9uUvbRSjFowFdQFumnVlKNfAlhKOQFOZpifFIwRFYda8lzvlJv1CzHEt500HgL2qofoIOTzFQNeJ_XkOQvRBy4eBkwxKvbHlwUAObxzZrCBjaAeQRGrMU926zpujSFQ_9KzUqNsNrxJWkBybOFViQp5mMZGFIWJbdt_oiROwZLG-NDK2i932hepUfr0i52mrTX-M9vTwy4uQsiMh2eSI7Ntghw0_xgrqqp6HZON7RPdKo2ldC5_rt9TFKKmyXvhZFLgxwsm8bzvqlIbV4KwNbEZIhh-n0")

	p := New()
	p.KubeConfig = kubeConf

	// Test with a valid pub key
	pubkey, err := crypto.ParseRSAPublicKeyFromPEM(validRS256pubkey)
	if err != nil {
		t.Fatal(err)
	}

	p.Use(
		WithJWT(),
		WithRS256(RS256Config{PublicKey: pubkey}),
	)(p).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("Received status code '%d'. Response: '%s'", status, rr.Body.String())
	}

	expected := string(`{"apiVersion":"v1","items":[],"kind":"List","metadata":{"resourceVersion":"","selfLink":""}}`)
	assert.JSONEq(expected, rr.Body.String(), "Got unexpected response body")

	// Test with an invalid pub key
	rr = httptest.NewRecorder()
	pubkey, err = crypto.ParseRSAPublicKeyFromPEM(invalidRS256pubkey)
	if err != nil {
		t.Fatal(err)
	}

	p.Use(
		WithJWT(),
		WithRS256(RS256Config{PublicKey: pubkey}),
	)(p).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Fatalf("Received status code '%d'. Response: '%s'", status, rr.Body.String())
	}

	expected = "crypto/rsa: verification error\n"
	assert.Equal(expected, rr.Body.String(), "Got unexpected response body")

}
