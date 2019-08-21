package multikube

import (
	"crypto/x509"
)

type Config struct {
	OIDCIssuerURL  string
	RS256PublicKey *x509.Certificate
	JWKS           *JWKS
}
