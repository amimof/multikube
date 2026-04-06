
# TLS Certificates

Multikube does not support insecure non-tls communication from clients so TLS certificates are required to run Multikube. By default, multikube generates a server private/public x509 certificate pair on startup. In most cases this is adequate however it's strongly recommended to provide a pair of certificates issued by a trusted CA.

<!--toc:start-->
- [TLS Certificates](#tls-certificates)
  - [Server Certificates](#server-certificates)
    - [Generate CA Certificates](#generate-ca-certificates)
    - [Generate a Server Certificate](#generate-a-server-certificate)
  - [Configure multikube to Use TLS](#configure-multikube-to-use-tls)
  - [Configure multikubectl](#configure-multikubectl)
<!--toc:end-->

## Server Certificates

> Note: Keep private keys (`ca.key`, `server.key`) secure and limit access to them.
>
### Generate CA Certificates

1. Openssl configuration file

   We'll be using a openssl configuration file for generating the CA cert. You can use the following as a starting point:

   ```toml {filename="ca.conf"}
   [ req ]
   default_bits        = 4096
   prompt              = no
   default_md          = sha256
   distinguished_name  = req_distinguished_name
   x509_extensions     = v3_ca  # The extension to add when creating a CA
 
   [ req_distinguished_name ]
   C  = SE
   ST = Halland
   L  = Varberg
   O  = multikube
   OU = multikube
   CN = multikube-ca
 
   [ v3_ca ]
   subjectAltName = @alt_names
   basicConstraints = CA:TRUE
   keyUsage = keyCertSign, cRLSign
   subjectKeyIdentifier = hash
   authorityKeyIdentifier = keyid:always,issuer
 
   [ alt_names ]
   DNS.1 = localhost
   IP.1  = 127.0.0.1
   ```

2. Generate a private key

   ```bash
   openssl genrsa -out ca.key 2048
   ```

3. Generate a self-signed CA certificate

   ```bash
   openssl req -x509 -new -days 365 -sha256 -nodes \
     -key ca.key \
     -out ca.crt \
     -config ca.conf

   ```

This will create `ca.key` and `ca.crt` in the current directory.

### Generate a Server Certificate

1. Openssl configuration file

   ```toml {filename=server.conf}
   [ req ]
   default_bits       = 2048
   prompt             = no
   default_md         = sha256
   distinguished_name = req_distinguished_name
   req_extensions     = v3_req
   
   [ req_distinguished_name ]
   C  = SE
   ST = Halland
   L  = Varberg
   O  = multikube
   OU = multikube
   CN = multikube
   
   [ v3_req ]
   basicConstraints = CA:FALSE
   keyUsage = digitalSignature, keyEncipherment
   extendedKeyUsage = serverAuth, clientAuth
   subjectAltName = @alt_names
   
   [ alt_names ]
   DNS.1 = localhost
   DNS.2 = multikube
   IP.1  = 127.0.0.1
   ```

2. Generate a server private key

   ```bash
   # Generate a 2048-bit server private key
   openssl genrsa -out server.key 2048
   ```

3. Generate a certificate signing request (CSR) using server.conf

   ```bash
   openssl req -new -sha256 \
     -key server.key \
     -out server.csr \
     -config server.conf
   ```

4. Sign the CSR with the local CA to produce server.crt

   ```bash
   openssl x509 -req -sha256 -days 365 \
     -in server.csr \
     -CA ca.crt \
     -CAkey ca.key \
     -CAcreateserial \
     -out server.crt \
     -extfile server.conf \
     -extensions v3_req
   ```

This will create `server.key`, `server.csr`, `server.crt`, and (if not present) `ca.srl`.

## Configure multikube to Use TLS

You can now start multikube by passing you newly created certificates as command line flags

```bash
multikube \
  --tls-key server.key \
  --tls-certificate server.crt
```

## Configure multikubectl

The multikube control plane should now have valid certificate configuration. Clients can use the CA to verify the server certificate instead of passing the `--insecure` flag.

```bash
multikubectl config create-server prod-cluster --address localhost:5743 --tls --ca ca.crt
```

> Make sure to change `localhost` to a host that multikube listens on
