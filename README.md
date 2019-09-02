# multikube
[![Build Status](https://travis-ci.org/amimof/multikube.svg?branch=master)](https://travis-ci.org/amimof/multikube)

Multikube is a multi-cluster API router for [Kubernetes](http://kubernetes.io/). It is responsible of routing requests to one or more Kubernetes API's (`kube-apiserver`) based on client tokens. Multikube acts as a proxy sitting between clients, such as `kubectl`, `curl`, web apps etc, and Kubernetes kube-apiserver thus separating the network into two security domains.

Features
* Off-loads API calls by re-using connections and serving data from cache
* Split network security domains
* Total transparency means compatibility with any kubectl command
* Minimal configuration required
* Audit logs 
* No database or dependencies makes scaling multikube a breeze 
* Configured with a kubeconfig

## How it works

In it's core, Multikube is a special type of HTTP router capable of routing requests based on the `HTTP authorization header`. Multikube expects clients to send special *"multikube-friendly"* access tokens when they wish to leverage the routing capabilities of Multikube. These special access tokens are JWT's ([JSON Web Tokens](https://jwt.io/)), much like those in Kubernetes, but with additional metadata that tells Multikube which kube-apiserver to route the request to.

A typical Multikube JWT header might look like this. The following shows a token with two important fields that are necessary in order for multikube to successfully route the request to the correct Kubernetes API. In this case a cluster named *minikube*. These fields are `sub` and `ctx`. 
```json
{
  "sub": "developer",
  "name": "John Doe",
  "admin": true,
  "iat": 1516239022,
  "ctx": "minikube"
}
```
View the entire token on [jwt.io](https://jwt.io/#debugger-io?token=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJkZXZlbG9wZXIiLCJuYW1lIjoiSm9obiBEb2UiLCJhZG1pbiI6dHJ1ZSwiaWF0IjoxNTE2MjM5MDIyLCJjdHgiOiJtaW5pa3ViZSJ9.QlpKTHk4atmRvQ_1BSLN9V4qJTUo-3FL8Ep3i5DJH_s2fSE8F6ZKFGma5DJr-Owmkla0xo5Nv9rf-b8UfotDXpU2cz4mhFNIj23SPLlzP4HJNOkRCZbJH89qm-5ny4-fpv_H56IMBrAyesyEt79HnNC1y8BJtMvcaJBxP5ufWRcl0CmGtEJceKRNWnh_qRJ5hjHjkEPdRBx5BsggSkYmL5tJXw5KBkLXvLlppN72TsPV9pjb3gbl6z_FPUyGutRdedFoOEIB8hHPKO-mTBymm0royjURDrY6jVzOvz9empLlO0RGV9AxKCoWz_eHvXBdCcYOyZAy2KcGHyvkAZMTPA)

The `sub` field corresponds to the user which Multikube will impersonate all API calls to Kubernetes as. By setting the `Impersonate-User` header. The `ctx` field corresponds to the `context` name configured in the kubeconfig file provided to Multikube on the command-line during startup. Clients using this JWT will ask Multikube to try route requests to the context minikube as user developer. 

If the value of `ctx` is set to something that multikube doesn't recognize (not configured in the kubeconfig), then you will get a `ContextNotFound` error. Multikube doesn't care about what the value of `sub` is. Just as long as rolebinding exists on that Kubernetes cluster. And if it doesn't then that user isn't granted any permissions. In other terms, Multikube does not apply any authorization policies, Kubernetes does.

## Connecting Multikube to Kubernetes clusters

Multikube communicates with Kubernetes clusters over separate TCP connections than those established from clients to Multikube. This means that connections can be re-used and shared for better performance. It also means that client connections are never used to communicate with the Kubernetes API. 

You define the set of Kubernetes clusters through the well-known [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) file. 

```
kubeconfig config set-cluster minikube \
  --server=https://192.168.99.100:8443 \
  --certificate-authority=ca.crt \
  --embed-certs=true \
  --kubeconfig=/etc/multikube/multikube.kubeconfig

kubectl config set-credentials minikube \
  --client-certificate=client.crt \
  --client-key=client.key \
  --embed-certs=true \
  --kubeconfig=/etc/multikube/multikube.kubeconfig

kubectl config set-context minikube \
  --cluster=minikube \
  --user=minikube \
  --kubeconfig=/etc/multikube/multikube.kubeconfig
```

With this kubeconfig file, clients having a JWT where the `ctx` field in the header is set to minikube will be routed to the cluster minikube. Ofcourse the kubeconfig file may contain as many contexts as you like. Note that we haven't set the `current-context` (kubectl config use-context) because Multikube never uses that field. It just compares the `ctx` field of client JWT's with the list of contexts in the provided kubeconfig file and tries to find a match.  

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
  --scheme=https \
  --tls-port=8443 \
  --tls-certificate=server.pem \
  --tls-key=server-key.pem \
  --tls-signer-certificate=/etc/multikube/signer.pem \
  --kubeconfig=/etc/multikube/multikube.kubeconfig
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
  --scheme=https \
  --tls-port=8443 \
  --tls-ca=ca.pem \
  --tls-certificate=server.pem \
  --tls-key=server-key.pem \
  --tls-signer-certificate=/etc/multikube/signer.pem \
  --kubeconfig=/etc/multikube/multikube.kubeconfig
```

## Non-TLS example (not recommended)
This method of serving multikube is not recommended for obvious reasons. Additionally, this method prevents kubectl compatibility due to how kubectl discards the `token` field in kubectl when the cluster has a http scheme over https.

```
multikube \
  --scheme=http \
  --port=8080 \
  --tls-signer-certificate=/etc/multikube/signer.pem \
  --kubeconfig=/etc/multikube/multikube.kubeconfig
```

## Generating a JWT for clients

##