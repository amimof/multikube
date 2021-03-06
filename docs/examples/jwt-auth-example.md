# Mutual Authentication

Multikube supports mutual authentication using JWT RS256 signing method. Multikube is to be configured with a RSA public key and the JWT needs to be issued using the private key. The JWT token from clients is then verified by multikube using the public key.

## Generate RSA key pairs

Use `openssl` to create a key pair

```
openssl genrsa -out rsa.key 4096
openssl rsa -in rsa.key -pubout > rsa.key.pub
```

## Sign a JWT 

Generate a signed JWT token using the rsa private key created previously. There are many ways of doing this but the fastest is to use https://jwt.io/.

1. Browse to https://jwt.io
2. Change algorithm to RS256
3. Change the `sub` claim in the payload to a desired username
4. Paste the content of rsa.key.pub in the public key text box
5. Paste the content of rsa.key in the private key text box

The JWT access token generated by jwt.io can now be used by any HTTP client to make requests to multikube. 

## Run Multikube

Now we can run multikube with RS256 validation enabled using the flag `--rs256-public-key`. If you haven't already, follow [this guide](https://github.com/amimof/multikube/blob/master/docs/examples/tls-example.md) in order configure TLS since it's required for mutual auth. 

```
multikube \
  --tls-certificate=server.pem \
  --tls-key=server-key.pem \
  --rs256-public-key=rsa.key.pub
```