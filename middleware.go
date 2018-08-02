package multikube

import (
	"log"
	"context"
	"net/http"
	"crypto/x509"
	"encoding/pem"
	"github.com/SermoDigital/jose/jws"
	"github.com/SermoDigital/jose/crypto"
)

type MiddlewareFunc func(next http.HandlerFunc) http.HandlerFunc
type Middleware func(http.Handler) http.Handler

// WithEmpty is an empty handler that does nothing
func WithEmpty(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Middleware (just a http.Handler)
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s %s %s", r.Method, r.URL.Path, r.URL.RawQuery, r.RemoteAddr, r.Proto)
		next.ServeHTTP(w, r)
	})
}

func WithValidate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		pub := `-----BEGIN CERTIFICATE-----
MIICwjCCAaoCCQC0W/8VU0Av9DANBgkqhkiG9w0BAQsFADAjMSEwHwYDVQQDDBhy
b2JlcnRzLW1hY2Jvb2subWRsd3Iuc2UwHhcNMTgwNTI4MTUwMjM0WhcNMTkwNTI4
MTUwMjM0WjAjMSEwHwYDVQQDDBhyb2JlcnRzLW1hY2Jvb2subWRsd3Iuc2UwggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDdG3G7UuxguZab8bBXS0aKtQck
jieC4D5+ap+jiZmnRU/9rP/pFjFtzBqlihCNBnXOhcf/iT6pXPP74FxPKR8wM7cK
M7V0r4q8mpEfM3nwGtwnAAPT8GWebchyatXKIGK+gjWHKaPZlkLrpxb3gwOEQbHD
hULkTH6qapwbO78Bn7oQCqPWBvUKlVQn1/K2BJmaH/6fy4jyqU0r0ABzqEPcLhrC
KUWYxX2NIbVtVkvR7CHjjSS7KYPZhW7EELrr1U4MkADcnGbI0/wYwjedEaHAR4mZ
N+JJvTNGI3V4NJYXb7ZRz8InIMMCEf3W2s9fn3K9aouhNFAl6xb9PXZQsucjAgMB
AAEwDQYJKoZIhvcNAQELBQADggEBAFtk8xXGPVXLSGtL7dDrFJXV8bIxoWmvX0mL
rv8RQxf42SKgfxJL0ASMNPZgLaLuiYhXWo7dWiDO+xOLYpuPxJGt10e2xK5P0cIf
CGO+8c6oK1BEDVK9zsVahSPcRaskPhNLgEg2/O5Se5oWaZKlpyMf2W+MyUyg/nDX
xfsYWjMZzqE1viwGB/Ay+KI7eWOlYenbX+1q2bh0pHX+ecWp+QsXNv6D9qA8TJoY
PVUFvSC/kkMUDUPUJH5veC+DCasU/ydhGIMxo2ThJJ1FcuQNaIiP5+VK1MGUQewd
77A3FULJ/I+ozUIYid+ygs/qHeP7zuBR8ArKU5hGagUOMVdaOC8=
-----END CERTIFICATE-----`
		
		block, _ := pem.Decode([]byte(pub))

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			log.Printf("ERROR Parse Cert: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		t, err := jws.ParseJWTFromRequest(r)
		if err != nil {
			log.Printf("ERROR token: %s", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if t == nil {
			log.Printf("ERROR token: No token")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		err = t.Validate(cert.PublicKey, crypto.SigningMethodRS256)
		if err != nil {
			log.Printf("ERROR validate: %s", err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "Target", t.Claims().Get("target").(string))
		ctx = context.WithValue(ctx, "Subject", t.Claims().Get("sub").(string)) 

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
