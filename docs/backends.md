# Backends

<!--toc:start-->
- [Backends](#backends)
  - [What a backend is](#what-a-backend-is)
  - [When to use a backend](#when-to-use-a-backend)
  - [How backends are used](#how-backends-are-used)
  - [Basic example](#basic-example)
  - [Example scenarios](#example-scenarios)
  - [Create a backend with the CLI](#create-a-backend-with-the-cli)
  - [Current behavior and caveats](#current-behavior-and-caveats)
  - [Full Config](#full-config)
<!--toc:end-->

A backend defines how Multikube connects to one upstream Kubernetes API server.

## What a backend is

A backend contains the connection details for one upstream cluster:

- the upstream API server URL
- the CA to trust when verifying the upstream server
- the credentials to use when authenticating to the upstream server
- optional TLS verification overrides
- an optional cache TTL

Routes point to backends with `backend_ref`. Policies then evaluate requests after a route has selected a backend.

## When to use a backend

Create a backend when you want Multikube to send requests to:

- a single Kubernetes cluster behind a stable API server URL
- several clusters, each with their own credentials and CA
- a development cluster with self-signed TLS
- a cluster that requires bearer token, basic auth, or client certificate authentication

## How backends are used

At runtime Multikube uses the backend to build the outgoing connection to the upstream API server.

- `server` becomes the target URL for forwarded requests
- `ca_ref` loads a certificate authority resource and uses it as the TLS root CA pool
- `auth_ref` loads a credential resource and applies it to the upstream request or TLS client config
- `insecure_skip_tls_verify` disables certificate verification for the upstream connection
- `cache_ttl` is parsed as a duration and stored on the backend runtime

## Basic example

```yaml
version: backend/v1
meta:
  name: prod-cluster
config:
  name: prod-cluster
  server: https://prod-api.example.internal:6443
  ca_ref: prod-ca
  auth_ref: prod-token
  insecure_skip_tls_verify: false
  cache_ttl: 30s
```

This backend tells Multikube to:

- connect to `https://prod-api.example.internal:6443`
- verify the server certificate with the `prod-ca` CA resource
- authenticate with the `prod-token` credential resource
- keep TLS verification enabled
- use a cache TTL of `30s`

## Example scenarios

### Token-authenticated cluster

Use a backend with `auth_ref` pointing to a credential that contains a bearer token.

```yaml
version: backend/v1
meta:
  name: staging
config:
  name: staging
  server: https://staging-api.example.internal:6443
  ca_ref: staging-ca
  auth_ref: staging-token
```

This is the most common setup when the upstream API server accepts bearer tokens.

### Client certificate authentication to the upstream cluster

Use a backend with `auth_ref` pointing to a credential that references a certificate resource.

```yaml
version: backend/v1
meta:
  name: mtls-cluster
config:
  name: mtls-cluster
  server: https://mtls-api.example.internal:6443
  ca_ref: mtls-ca
  auth_ref: mtls-client-credential
```

In this flow:

- the CA verifies the upstream server certificate
- the credential provides a client certificate for mutual TLS

### Development cluster with self-signed or broken TLS

```yaml
version: backend/v1
meta:
  name: dev-cluster
config:
  name: dev-cluster
  server: https://dev-api.example.internal:6443
  insecure_skip_tls_verify: true
```

This can be useful during local testing, but it disables upstream certificate verification and should not be used in normal production setups.

## Create a backend with the CLI

```bash
multikubectl create backend prod-cluster \
  --server https://prod-api.example.internal:6443 \
  --ca-ref prod-ca \
  --auth-ref prod-token \
  --cache-ttl 30s
```

Useful flags:

- `--server` required upstream API server URL
- `--ca-ref` reference to a CA resource
- `--auth-ref` reference to a credential resource
- `--insecure-skip-tls-verify` disable upstream certificate verification
- `--cache-ttl` duration such as `30s`, `5m`, or `1h`
- `--label` attach one or more metadata labels

## Current behavior and caveats

- `ca_ref` must point to an existing CA resource or compilation fails
- `auth_ref` must point to an existing credential resource or compilation fails
- if the credential referenced by `auth_ref` points to a client certificate, that certificate must exist or compilation fails
- backend health is present in status, but the compiler currently does not skip unhealthy backends
- `cache_ttl` is optional and defaults to zero when omitted
- `server` is required in practice; parsing errors are caught during compilation

## Full Config

```yaml
version: backend/v1 # Required. API version for backend resources.
meta:
  name: prod-cluster # Required. Resource name. Must be unique for this backend resource type.
  labels: # Optional. Arbitrary metadata labels used for organization and filtering.
    env: production
    team: platform
  created: "2026-04-04T12:00:00Z" # Server-managed. Creation timestamp.
  updated: "2026-04-04T12:00:00Z" # Server-managed. Last update timestamp.
  generation: 1 # Server-managed. Monotonic version of desired state.
  resource_version: 1 # Server-managed. Internal resource revision.
  uid: "11111111-2222-3333-4444-555555555555" # Server-managed. Unique identifier.
config:
  name: prod-cluster # Optional but normally set to the same value as meta.name.
  server: https://prod-api.example.internal:6443 # Required in practice. Full upstream Kubernetes API URL.
  ca_ref: prod-ca # Optional. Name of a CA resource used to verify the upstream server certificate.
  auth_ref: prod-token # Optional. Name of a credential resource used for upstream auth.
  insecure_skip_tls_verify: false # Optional. Defaults to false. When true, disables upstream TLS certificate verification.
  cache_ttl: 30s # Optional. Duration string. Examples: 0s, 30s, 5m, 1h.
status:
  healthy: true # Runtime status field. Not currently used to exclude a backend from compilation.
```

Field notes:

- `version` should be `backend/v1`
- `meta.name` is the stable identifier other resources use when they refer to this backend
- `config.name` is part of the schema and is typically set to the same value as `meta.name`
- `server` should include scheme, host, and port
- `ca_ref` should be set when the upstream server uses TLS and you want normal certificate verification
- `auth_ref` should reference a credential that matches the upstream server's auth method
- `insecure_skip_tls_verify: true` trades safety for convenience and is mainly for development or debugging
