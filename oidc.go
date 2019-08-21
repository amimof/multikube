package multikube

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type JWKS struct {
	Keys []JSONWebKey `json:"keys"`
}

type JSONWebKey struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

type OpenIDConfiguration struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
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
func getWellKnown(u string) (*OpenIDConfiguration, error) {

	client := &http.Client{}
	wellKnownUrl := fmt.Sprintf("%s/%s", u, "/.well-known/openid-configuration")
	req, err := http.NewRequest("GET", wellKnownUrl, nil)
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

	var c *OpenIDConfiguration
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}

	return c, nil

}

// GetJWKSFromURL fetches the keys of an OpenID Connect endpoint.
// The endpoint must support OpenID Connect discovery as per https://openid.net/specs/openid-connect-discovery-1_0.html
func GetJWKSFromURL(u string) (*JWKS, error) {

	// Make a request and fetch content of .well-known url (http://some-url/.well-known/openid-configuration)
	c, err := getWellKnown(u)
	if err != nil {
		return nil, err
	}

	// Get content of jwks_keys field
	j, err := getKeys(c.JWKSURI)
	if err != nil {
		return nil, err
	}
	return j, nil
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
