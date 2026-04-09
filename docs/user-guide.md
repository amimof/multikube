# User Guide

<!--toc:start-->
- [User Guide](#user-guide)
  - [Prerequisites](#prerequisites)
  - [1: Create two local clusters with kind](#1-create-two-local-clusters-with-kind)
  - [2: Add something unique to each cluster](#2-add-something-unique-to-each-cluster)
  - [3: Import both clusters](#3-import-both-clusters)
  - [4: Create routes](#4-create-routes)
  - [5: Create a test token](#5-create-a-test-token)
  - [6: Create policies](#6-create-policies)
  - [7: Test routing with curl](#7-test-routing-with-curl)
    - [7a. Route to the east cluster](#7a-route-to-the-east-cluster)
    - [7b. Route to the west cluster](#7b-route-to-the-west-cluster)
    - [7c. List namespaces through either route](#7c-list-namespaces-through-either-route)
  - [8: Test policy enforcement with curl](#8-test-policy-enforcement-with-curl)
  - [9: Try a request without a token](#9-try-a-request-without-a-token)
  - [Clean up](#clean-up)
<!--toc:end-->

This guide walks through a small local demo that shows how Multikube works with two `kind` Kubernetes clusters. You will learn how to create backends, routes, polices and use `cURL` to verify routing and policy behavior.

This guide does not cover how to install `multikube` or `multikubectl`. Follow the [installation docs](/docs/README.md) first, then come back here.

## Prerequisites

Make sure these are already available:

- `kind`
- `kubectl`
- `curl`

You also need a running Multikube control plane and a working `multikubectl` config. Read the [getting started guide](/docs/getting-started.md) to learn how to set it up.

If you are running multikube in Docker then you must join the multikube container to the kind network. Otherwise multikube will not be able to communicate with the kubernetes clusters. If multikube is already running then you can run this command to join multikube to the `kind` network:

```bash
docker network connect kind multikube
```

If multikube is not setup and you want to run it with docker, use `--network kind` flag

```bash
docker run -d \
  --name multikube \
  --network kind  \
  -p 5743:5743 \
  -p 8443:8443 \
  -v multikube-data:/.local/state/multikube  \
  multikube:latest
```

## 1: Create two local clusters with kind

Create two clusters named `demo-east` and `demo-west`:

```bash
kind create cluster --name demo-east
kind create cluster --name demo-west
```

`kind` adds both clusters to your kubeconfig automatically.

List the contexts to confirm they exist:

```bash
kubectl config get-contexts
```

You should see two contexts named: `kind-demo-east` and `kind-demo-west`

## 2: Add something unique to each cluster

Create one namespace in each cluster so it is easy to see which cluster answered your request later.

```bash
kubectl --context kind-demo-east create namespace east-demo
kubectl --context kind-demo-west create namespace west-demo
```

## 3: Import both clusters

The `import` command is very useful when you want to create multikube resources from a kubeconfig. It reads the provided kubeconfig and imports a `context` into multikube creating `backend`, `certificate`, `ca` and `credentials` resource.

> A note on kind: Multikube accesses kind clusters through the docker network so we can't import the kubeconfig that kind created when creating the clusters. To create kubeconfigs that can be used from inside other docker containers we have to use the `--internal` flag.

Create an `internal` kubeconfig file per cluster that we import into multikube later

```bash
kind get kubeconfig --internal --name demo-east > ~/.kube/demo-east.yaml
kind get kubeconfig --internal --name demo-west > ~/.kube/demo-west.yaml
```

Import the two `kind` contexts and give them easy-to-remember backend names:

```bash
multikubectl import kind-demo-east --backend-name east-cluster --kubeconfig ~/.kube/demo-east.yaml
multikubectl import kind-demo-west --backend-name west-cluster --kubeconfig ~/.kube/demo-west.yaml
```

> If you re-run the guide and the resources already exist, use `--force` to overwrite existing resources

List the imported backends:

```bash
multikubectl get backends
```

You should see both `east-cluster` and `west-cluster`.

## 4: Create routes

Create one route for each backend using the `X-Cluster` header:

```bash
multikubectl create route route-east \
  --backend-ref east-cluster \
  --header-name X-Cluster \
  --header-value east

multikubectl create route route-west \
  --backend-ref west-cluster \
  --header-name X-Cluster \
  --header-value west
```

> Multikube supports multiple routing mechanisms such as JWT, PathPrefix and Header

Verify that the routes exist:

```bash
multikubectl get routes
```

If routing compiled successfully, the routes should move to `READY`.

## 5: Create a test token

Create a token for a user in the `demo-readonly` group:

```bash
TOKEN=$(multikubectl create token \
  --subject demo-user \
  --group demo-readonly \
  --ttl 1h)
```

You can inspect the token variable with:

```bash
echo "${TOKEN}"
```

## 6: Create policies

We're going to create two policy resources. One allow policy for read-only access and one deny polcy for deletes. At the moment, `multikubectl create policy` creates empty shell policies. The command does not expose flags for building detailed rules. For this guide, use `apply` to create policies from yaml manifests.

```yaml {filename="demo-policies.yaml"}
---
version: policy/v1
meta:
  name: demo-readonly
config:
  name: demo-readonly
  rules:
    - effect: EFFECT_ALLOW
      subjects:
        - groups:
            - demo-readonly
      clusters:
        - names:
            - east-cluster
            - west-cluster
      resources:
        - resource: namespaces
      actions:
        - ACTION_GET
        - ACTION_LIST
---
version: policy/v1
meta:
  name: demo-deny-delete
config:
  name: demo-deny-delete
  rules:
    - effect: EFFECT_DENY
      subjects:
        - groups:
            - demo-readonly
      clusters:
        - names:
            - east-cluster
            - west-cluster
      resources:
        - resource: namespaces
      actions:
        - ACTION_DELETE
```

Create the two policies with this command

```bash
multikubectl apply -f demo-policies.yaml
```

## 7: Test routing with curl

Now send requests through the Multikube proxy.

### 7a. Route to the east cluster

This request should return the `east-demo` namespace object from the east cluster:

```bash
curl -sk \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-Cluster: east" \
  https://127.0.0.1:8443/api/v1/namespaces/east-demo
```

You should see JSON containing `east-demo`.

### 7b. Route to the west cluster

This request should return the `west-demo` namespace object from the west cluster:

```bash
curl -sk \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-Cluster: west" \
  https://127.0.0.1:8443/api/v1/namespaces/west-demo
```

You should see JSON containing `west-demo`.

### 7c. List namespaces through either route

```bash
curl -sk \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-Cluster: east" \
  https://127.0.0.1:8443/api/v1/namespaces

curl -sk \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-Cluster: west" \
  https://127.0.0.1:8443/api/v1/namespaces
```

The two responses should both be valid Kubernetes API responses, but they should reflect different backend clusters.

## 8: Test policy enforcement with curl

Try to delete one of the namespaces:

```bash
curl -sk -i \
  -X DELETE \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "X-Cluster: east" \
  https://127.0.0.1:8443/api/v1/namespaces/east-demo
```

The request should be denied.

In most setups you should see an HTTP `403 Forbidden` response, because the deny policy matches before any allow rule can permit the request.

## 9: Try a request without a token

Policies are evaluated against the authenticated principal. If you omit the bearer token, the request should not be authorized.

```bash
curl -sk -i \
  -H "X-Cluster: east" \
  https://127.0.0.1:8443/api/v1/namespaces
```

You should get an authentication or authorization failure.

## Clean up

Delete the local `kind` clusters when you are done:

```bash
kind delete cluster --name demo-east
kind delete cluster --name demo-west
```

If you also want to remove the imported resources from Multikube, delete them with `multikubectl`:

```bash
multikubectl delete route route-east
multikubectl delete route route-west
multikubectl delete policy demo-readonly
multikubectl delete policy demo-deny-delete
multikubectl delete backend east-cluster
multikubectl delete backend west-cluster
```
