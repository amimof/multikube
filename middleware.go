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
MIICvDCCAaQCCQC7lkJGUK5yrTANBgkqhkiG9w0BAQsFADAgMQswCQYDVQQGEwJT
RTERMA8GA1UEAwwIbWluaWt1YmUwHhcNMTgwODAyMTM1NjI0WhcNMTkwODAyMTM1
NjI0WjAgMQswCQYDVQQGEwJTRTERMA8GA1UEAwwIbWluaWt1YmUwggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQC2vGjj3Frrcl8WVoJVlpSt2TPda64+9jyV
rVJiGcMhnY2ZbepmAKWadHG/euHwkYc2TmGzwgm2vhPSbaYyrGE5GMwOSnzX2Lwj
rCZr9PzijD05o3ci61VB5B/9PTgyNXOYljmqmGFNOhr2CGU4ifn+3IMHbc4cEr3M
AAkLTbAENk06kVYlR2S297nPbyAd9/nzR5uR9aQ+46TAS530zHL5Nvb8XA9VVc/y
Z2704URrRu2rc/1Lzz36nRayXFRux/JTUylZISenF5pfrUBN2gI55no/R1zm0NMD
p1wuu1ou4IOBsgZIOlsJHglmT0A0JMFBllqtYN9AMuU2nyOE6UpVAgMBAAEwDQYJ
KoZIhvcNAQELBQADggEBAFvPbN6IZ26gCVy3BTxJ+3Cc6R9VDS2yH86wP9iXQ/87
7+yg+u+H9oTCZOkR/Jt7UB/3oDy4IhMfv2ysh3+v3+FbgXc44WuCDYHmFTZB/o+G
4YR0/58rLLU6w5We8BduICzT32fmbkaAV9NObO4cXooDwLzV9Tiwjk83avaMMawg
Y/hnV1dSgVTfL/VMOahi030PgV+EsrmOXA1u424YmF6Xrc0MlWZadMUitdns46Pl
HuHLMeKcrWw6wPVnysSeJeu+RsNxuW0LBroLpcvRYrszEeHwrN/ljy1/PJFW6n6V
J5xPObhvjG6pisQUWwSRYm7B5dH7hLbOT3pMTjlFscc=
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

		ctx := context.WithValue(r.Context(), "Context", t.Claims().Get("ctx").(string))
		ctx = context.WithValue(ctx, "Subject", t.Claims().Get("sub").(string)) 

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
