package multikube

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// Config holds a top-level configuration of an instance of Multikube. It is used to
// pass around configuration used by different packages within the project.
type Config struct {
	OIDCIssuerURL    string
	OIDCPollInterval time.Duration
	RS256PublicKey   *x509.Certificate
	JWKS             *JWKS
}

// JWKS is a representation of Json Web Key Store. It holds multiple JWK's in an array
type JWKS struct {
	Keys []JSONWebKey `json:"keys"`
}

// JSONWebKey is a representation of a Json Web Key
type JSONWebKey struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// openIDConfiguration is an internal type used to marshal/unmarshal openid connect configuration
// from the provider.
type openIDConfiguration struct {
	Issuer  string `json:"issuer"`
	JwksURI string `json:"jwks_uri"`
}

// getTokenFromRequest returns a []byte representation of JWT from an HTTP Authorization Bearer header
func getTokenFromRequest(req *http.Request) []byte {
	if ah := req.Header.Get("Authorization"); len(ah) > 7 && strings.EqualFold(ah[0:7], "BEARER ") {
		return []byte(ah[7:])
	}
	return nil
}

// dials an url which returns an array of Json Web Keys. The URL is typically
// an OpenID Connect .well-formed URL as per https://openid.net/specs/openid-connect-discovery-1_0.html
// Unmarshals it's json content into JWKS and returns it
func getKeys(u string) (*JWKS, error) {

	client := &http.Client{}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jwks *JWKS
	err = json.Unmarshal(body, &jwks)
	if err != nil {
		return nil, err
	}

	return jwks, nil
}

// dials the .well-known url and unmarshals it's json content into an OpenIDConfiguration
// see https://openid.net/specs/openid-connect-discovery-1_0.html
func getWellKnown(u string) (*openIDConfiguration, error) {

	client := &http.Client{}
	wellKnownURL := fmt.Sprintf("%s/%s", u, "/.well-known/openid-configuration")
	req, err := http.NewRequest("GET", wellKnownURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var c *openIDConfiguration
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}

	return c, nil

}

// GetJWKSFromURL fetches the keys of an OpenID Connect endpoint in a go routine. It polls the endpoint
// every n seconds. Returns a cancel function which can be called to stop polling and close the channel.
// The endpoint must support OpenID Connect discovery as per https://openid.net/specs/openid-connect-discovery-1_0.html
func (c *Config) GetJWKSFromURL() func() {

	quit := make(chan int)
	go func() {
		for {
			time.Sleep(c.OIDCPollInterval)
			select {
			case <-quit:
				close(quit)
				return
			default:
				// Make a request and fetch content of .well-known url (http://some-url/.well-known/openid-configuration)
				w, err := getWellKnown(c.OIDCIssuerURL)
				if err != nil {
					log.Printf("ERROR retrieving openid-configuration: %s", err)
					continue
				}
				// Get content of jwks_keys field
				j, err := getKeys(w.JwksURI)
				if err != nil {
					log.Printf("ERROR retrieving JWKS from provider: %s", err)
					continue
				}
				c.JWKS = j
			}
		}
	}()

	return func() {
		quit <- 1
	}

}

// Find will loop through the keys on the JWKS and return that which has a matching key id
func (j *JWKS) Find(s string) *JSONWebKey {
	for _, v := range j.Keys {
		if s == v.Kid {
			return &v
		}
	}
	return nil
}
