package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"
)

// Config holds a top-level configuration of an instance of Multikube. It is used to
// pass around configuration used by different packages within the project.
type Config struct {
	OIDCIssuerURL          string
	OIDCUsernameClaim      string
	OIDCPollInterval       time.Duration
	OIDCInsecureSkipVerify bool
	OIDCCa                 *x509.Certificate
	RS256PublicKey         *x509.Certificate
	JWKS                   *JWKS
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

// dials an url which returns an array of Json Web Keys. The URL is typically
// an OpenID Connect .well-formed URL as per https://openid.net/specs/openid-connect-discovery-1_0.html
// Unmarshals it's json content into JWKS and returns it
func getKeys(u string, ca *x509.Certificate, i bool) (*JWKS, error) {

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := tlsClient(ca, i).Do(req)
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
// see https://openid.net/specs/openid-connect-discovery-1_0.html.
// Accepts a trusted CA certificate as well as a bool to skip tls verification
func getWellKnown(u string, ca *x509.Certificate, i bool) (*openIDConfiguration, error) {

	ul, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	ul.Path = path.Join(ul.Path, ".well-known/openid-configuration")

	//wellKnownURL := fmt.Sprintf("%s/%s", u, "/.well-known/openid-configuration")
	req, err := http.NewRequest("GET", ul.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := tlsClient(ca, i).Do(req)
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

// Creates an http client with TLS configuration. If ca is nil then client without TLS configuration is returned instead
// Set i to true to skip tls verification for this client
func tlsClient(ca *x509.Certificate, i bool) *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	// Add tls config to client if ca isn't nil
	if ca != nil {
		caPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Raw})
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		rootCAs.AppendCertsFromPEM(caPem)
		tlsConfig.RootCAs = rootCAs
	}

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

}

// GetJWKSFromURL fetches the keys of an OpenID Connect endpoint in a go routine. It polls the endpoint
// every n seconds. Returns a cancel function which can be called to stop polling and close the channel.
// The endpoint must support OpenID Connect discovery as per https://openid.net/specs/openid-connect-discovery-1_0.html
func (c *Config) GetJWKSFromURL() func() {

	// Make sure config has non-nil fields
	c.JWKS = &JWKS{
		Keys: []JSONWebKey{},
	}

	// Run a function in a go routine that continuously fetches from remote oidc provider
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
				w, err := getWellKnown(c.OIDCIssuerURL, c.OIDCCa, c.OIDCInsecureSkipVerify)
				if err != nil {
					log.Printf("ERROR retrieving openid-configuration: %s", err)
					continue
				}
				// Get content of jwks_keys field
				j, err := getKeys(w.JwksURI, c.OIDCCa, c.OIDCInsecureSkipVerify)
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
