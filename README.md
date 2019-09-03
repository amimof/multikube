# multikube
[![Build Status](https://travis-ci.org/amimof/multikube.svg?branch=master)](https://travis-ci.org/amimof/multikube) [![huego](https://godoc.org/github.com/amimof/multikube?status.svg)](https://godoc.org/github.com/amimof/multikube) [![Go Report Card](https://goreportcard.com/badge/github.com/amimof/multikube)](https://goreportcard.com/report/github.com/amimof/multikube)

Multikube is a modern HTTP reverse proxy for [Kubernetes](http://kubernetes.io/) API server. 

## Features
* Validates JSON Web Tokens (JWT) 
* OIDC (OpenID Connect) provider support
* Off-loads API calls by re-using connections and serving data from cache
* Split network security domains
* Total transparency means compatibility with any kubectl command
* Minimal configuration required
* Audit logs 
* Prometheus metrics
* No database or dependencies makes scaling multikube a breeze 
* Configured with a kubeconfig

## Overview

A client, wether it is kubectl, cURL or a browser, may make requests to Multikube as if it was a Kubernetes API. Multikube will validate the client access token and intelligently route the request to an upstream API-server, impersonating that user. The client can target a cluster, or *context*, using either a URL path or an HTTP header. 

Multikube communicates with Kubernetes clusters over separate TCP connections than those established from clients to Multikube. This means that connections can be re-used, shared and cached for better performance. It also means that client connections are never used to communicate with the Kubernetes API. 

As an example, `kubectl` uses a context that is configured to use the server `https://127.0.0.1:6443/k8s-dev-cluster` which happens to be Multikube running locally. Note the leading path in the URL which is the context name. Multikube will try to match that path with a context in it's kubeconfig. All traffic from kubectl will be routed through Multikube to the apiserver k8s-dev-cluster. 

## Getting started

Download the latest binary from the [release page](https://github.com/amimof/multikube/releases) for your target platform. Below is for Linux.
```
curl -LOs https://github.com/amimof/multikube/releases/latest/download/multikube-linux-amd64
``` 

Or use the official Docker scratch image
```
docker pull amimof/multikube:latest
```

## Configuration

You configure Multikube on the command line. There is no configuration file. However Multikube needs a [kubeconfig](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/) in order to communicate with upstream API servers. The kubeconfig is where you will configure the clusters (contexts) that Multikube will use for routing. 

## Examples

Examples are found under [docs/examples](https://github.com/amimof/multikube/blob/master/docs/examples/dex-example.md)

## Contributing

All help in any form is highly appreciated and your are welcome participate in developing together. To contribute submit a Pull Request. If you want to provide feedback, open up a Github Issue or contact me personally.