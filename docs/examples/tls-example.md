## TLS example (recommended)

Generate private keys
```
sudo openssl ecparam -name secp521r1 -genkey -noout -out ca-key.pem
sudo openssl ecparam -name secp521r1 -genkey -noout -out server-key.pem
```

Generate a CA certificate using the CA private key
```
sudo openssl req -x509 -new -sha256 -nodes -key ca-key.pem -out ca.pem -subj '/CN=multikube-ca' 
```

Generate the server certificate by creating a CSR and signing it with the CA key
```
sudo openssl req -new -sha256 -key server-key.pem -subj '/CN=localhost' -out server.csr
sudo openssl x509 -req -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out server.pem
```

Run multikube
```
multikube \
  --tls-certificate=server.pem \
  --tls-key=server-key.pem
```

## TLS example with mutual authentication 

Generate private key
```
sudo openssl ecparam -name secp521r1 -genkey -noout -out client-key.pem
```

Generate the client certificate by creating CSR and signing it with the CA key
```
sudo openssl req -new -sha256 -key client-key.pem -subj '/CN=multikube' -out client.csr
sudo openssl x509 -req -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out client.pem
```

Run multikube
```
multikube \
  --tls-ca=ca.pem \
  --tls-certificate=server.pem \
  --tls-key=server-key.pem
```