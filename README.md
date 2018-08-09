# multikube

Multikube is a multi-cluster API router for Kubernetes. It's responsible of routing requests to a designated `kube-apiserver` based on client tokens. Multikube acts as a proxy sitting between clients, such as `kubectl`, `curl`, web apps etc, and Kubernetes kube-apiserver thus separating the network into two security domains.

* Off-loads API calls by re-using connections and serving data from cache
* Split network security domains
* Total transparency means compatibility with any kubectl command
* Minimal configuration required
* Request interceptors allows for manipulating Kubernetes objects sent arount
* Audit logs 
* No database or dependencies makes scaling multikube a breeze 
* Configured with a kubeconfig from file or remote http server
* Supports JTI revocation lists in order to prevent rogue tokens

A typical multi-cluster Kubernetes topology
```
                                             +----------+
                                             |          |
                                             | Cluster1 |
          https://cluster1:443/              |          |
        +----------------------------------> +----------+
        |
        |                                    +----------+
        | https://kube.domain.io:8443        |          |
kubectl +----------------------------------> | Cluster2 |
        |                                    |          |
        |                                    +----------+
        |
        +----------------------------------> +----------+
          https://azmk8s.io:443              |          |
                                             | Cluster3 |
                                             |          |
                                             +----------+
```

With multikube
```
                                                            +----------+
                                                            |          |
                                                            | Cluster1 |
                             https://cluster1:443/          |          |
                           +------------------------------> +----------+
                           |
            +-----------+  |                                +----------+
            |           |  | https://kube.domain.io:8443    |          |
kubectl +---> Multikube +---------------------------------> | Cluster2 |
            |           |  |                                |          |
            +-----------+  |                                +----------+
                           |
                           +------------------------------> +----------+
                             https://azmk8s.io:443          |          |
                                                            | Cluster3 |
                                                            |          |
                                                            +----------+
```

## Why?


## How it works

In it's core, Multikube is a special type of HTTP router that is capable of routing requests based on either JWT access tokens or client certificates. For this to be possible Multikube expects clients to send special "multikube-friendly" access tokens when they wish to leverage the routing capabilities of multikube. These special access tokens are JWT's ([JSON Web Tokens](https://jwt.io/)), much like those in Kubernetes, but with additional metadata which tells multikube where to route the request. 
