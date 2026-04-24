# Backends

<!--toc:start-->
- [Backends](#backends)
  - [What a backend is](#what-a-backend-is)
  - [How backends are used](#how-backends-are-used)
  - [Basic example](#basic-example)
  - [Impersonation configuration](#impersonation-configuration)
  - [Probe configuration](#probe-configuration)
  - [Example scenarios](#example-scenarios)
    - [Multiple backends](#multiple-backends)
    - [Token-authenticated cluster](#token-authenticated-cluster)
    - [Client certificate authentication to the upstream cluster](#client-certificate-authentication-to-the-upstream-cluster)
    - [Development cluster with self-signed or broken TLS](#development-cluster-with-self-signed-or-broken-tls)
  - [Import a backend from kubeconfig](#import-a-backend-from-kubeconfig)
  - [Create a backend with the CLI](#create-a-backend-with-the-cli)
  - [Current behavior and caveats](#current-behavior-and-caveats)
  - [Full Config](#full-config)
<!--toc:end-->

API Reference: [backend/v1](https://github.com/amimof/multikube/blob/master/api/backend/v1/backend.proto)

## What a backend is

A backend defines how Multikube connects to a set of upstream Kubernetes API servers. Backends should be though of as kubernetes *clusters* but we use the backend terminology here since it makes more sence in context of load balancing.  A backend contains the connection details, impersionation and heartbeat probe configuration for one or more API servers. The configuration is shared and used by all servers part of the backend. In terms of proxy and routing, the servers are known as the backend pool. [Routes](/docs/usage/routes.md) point to backends with `backend_ref`. [Policie](/docs/usage/policies.md) then evaluate requests after a route has selected a backend.

## How backends are used

At runtime Multikube uses the backend to build the outgoing connection to the upstream API server. This is an important trust boundary: Multikube does not forward the end user's kubeconfig credentials directly to the backend cluster. Instead, each backend needs its own credentials that Multikube can use when it connects to that cluster. Those backend credentials are separate from the identity that the user presented to Multikube at the edge. 


In practice this means that users authenticate to Multikube. Multikube authorizes the request using its own policies then authenticates to the upstream Kubernetes API server with the credential configured on the backend. Those backend credentials must be created on the target cluster ahead of time. In most environments, a cluster administrator sets this up by creating a dedicated ServiceAccount and issuing a long-lived token for Multikube to use. That token, client certificate, or other supported secret material is then stored in a Multikube credential resource and referenced by the backend.

If you already have working cluster access in a kubeconfig file, [`multikubectl import`](#import-a-backend-from-kubeconfig) can speed this up by reusing the cluster CA and supported auth material from that kubeconfig context.

## Basic example

```yaml
version: backend/v1
meta:
  name: prod-cluster
config:
  servers:
  - https://prod-api.example.internal:6443
  ca_ref: prod-ca
  auth_ref: prod-token
  insecure_skip_tls_verify: false
```

This backend tells Multikube to:

- connect to `https://prod-api.example.internal:6443`
- verify the server certificate with the `prod-ca` CA resource
- authenticate with the `prod-token` credential resource
- keep TLS verification enabled

## Impersonation configuration

By default, Multikube uses impersonation configuration to perform requests on the upstream servers on behalf of the user. You can read more about it on the [official kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/user-impersonation/). In essence, impersonation is the ability for a user or service account to act as another user, group, or service account when making API requests, typically for debugging, auditing, or delegated access control. We use this mechanic to separate Multikubes credentials from the users credentials which is important.

> Note: Multikube does not provision RBAC on the clusters. It manages authorization and routing on the edge. Once a request is authorized in the Multikube realm, and successfully routed to a Kuberenets Cluster then authorization still occurs in that cluster. So platform engineers should, as always, provision appropriate RBAC rules for users accessing clusters. 

Backends can define `config.impersonation_config` to control how Multikube builds Kubernetes impersonation headers for upstream requests. Fields under `config.impersonation_config`:

- `name`: optional label for the impersonation profile
- `enabled`: enable or disable header injection
- `username_claim`: claim used for `Impersonate-User`
- `groups_claim`: claim used for `Impersonate-Group`
- `extra_claims`: claims copied into `Impersonate-Extra-<claim>` headers

If the `impersonation_config` field is omitted when creating a backend, a default impersonation configuration will be created automatically to ensure that impersionation always is enabled. A default impersionation configuration looks like this:

```yaml
version: backend/v1
meta:
  name: prod-cluster
config:
  servers:
  - https://prod-api.example.internal:6443
  ca_ref: prod-ca
  auth_ref: prod-token
  impersonation_config:
    name: default
    enabled: true
    username_claim: sub
    groups_claim: groups
```

Custom claim mapping example:

```yaml
version: backend/v1
meta:
  name: sso-cluster
config:
  servers:
  - https://sso-api.example.internal:6443
  ca_ref: sso-ca
  auth_ref: sso-backend-token
  impersonation_config:
    name: oidc
    enabled: true
    username_claim: email
    groups_claim: roles
    extra_claims:
    - tenant
    - scopes
```

## Probe configuration

Backends can define `config.probes` to control health and readiness checks for each upstream target in the backend pool. Multikube runs HTTP `GET` probes against each configured backend target. Probe results are reflected in backend status and runtime target eligibility.

Fields under `config.probes`:

- `healthiness`: health probe configuration (is it safe to send traffic?)
- `readiness`: readiness probe configuration (is it alive?)

Each probe supports:

- `path`: URL path to probe
- `timeout_seconds`: how long a single probe request may take
- `period_seconds`: how often the probe runs
- `failure_threshold`: consecutive failures before the target is marked unhealthy or not ready
- `success_threshold`: consecutive successes before the target is marked healthy or ready again
- `initial_delay_seconds`: delay before the first probe runs

Default behavior when `probes` is omitted:

- `healthiness.path: /healthz`
- `readiness.path: /readyz`
- `timeout_seconds: 1`
- `period_seconds: 5`
- `failure_threshold: 3`
- `success_threshold: 3`
- `initial_delay_seconds: 1`

Example:

```yaml
version: backend/v1
meta:
  name: prod-cluster
config:
  servers:
  - https://control-plane-1.example.internal:6443
  - https://control-plane-2.example.internal:6443
  probes:
    healthiness:
      path: /healthz
      timeout_seconds: 1
      period_seconds: 5
      failure_threshold: 3
      success_threshold: 3
      initial_delay_seconds: 1
    readiness:
      path: /readyz
      timeout_seconds: 1
      period_seconds: 10
      failure_threshold: 2
      success_threshold: 2
      initial_delay_seconds: 2
```

> Notes: probe failures include transport errors, timeouts, and any response other than `200 OK`. A target is excluded only when a configured probe has produced a known failing state. If probe state is still unknown, the target remains eligible for routing

## Example scenarios

### Multiple backends

Use multiple server URLs to let Multikube load balance requests across the backend pool.

> This feature is still evolving. Round robin is the default load balancing strategy today.

```yaml
version: backend/v1
meta:
  name: staging
config:
  servers:
  - https://control-plane-1.example.internal:6443
  - https://control-plane-2.example.internal:6443
  - https://control-plane-3.example.internal:6443
  ca_ref: staging-ca
  auth_ref: staging-token
```

### Token-authenticated cluster

Use a backend with `auth_ref` pointing to a credential that contains a bearer token.

```yaml
version: backend/v1
meta:
  name: staging
config:
  servers:
  - https://staging-api.example.internal:6443
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
  servers:
  - https://mtls-api.example.internal:6443
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
  servers:
  - https://dev-api.example.internal:6443
  insecure_skip_tls_verify: true
```

This can be useful during local testing, but it disables upstream certificate verification and should not be used in production.

## Import a backend from kubeconfig

If you already have working access to a Kubernetes cluster through a normal kubeconfig, the fastest way to connect that cluster to Multikube is `multikubectl import`.

The import command reads one kubeconfig context and creates the Multikube resources needed to talk to that cluster.

Depending on what is present in the kubeconfig, it can create:

- a backend
- a certificate authority resource from the cluster CA
- a credential resource for token or basic auth
- a certificate resource when the kubeconfig uses client certificate authentication

Basic usage:

```bash
multikubectl import prod
```

This reads the `prod` context from the default kubeconfig path and creates resources with names derived from the context:

- `prod-backend`
- `prod-credential`
- `prod-certificate`
- `prod-certificate-authority`

You can then verify the imported backend with:

```bash
multikubectl get backends
```

If your kubeconfig is not in the default location, pass it explicitly:

```bash
multikubectl import prod --kubeconfig /path/to/kubeconfig
```

You can also override the generated resource names:

```bash
multikubectl import prod \
  --backend-name prod-cluster \
  --credential-name prod-token \
  --certificate-name prod-client-cert \
  --certificate-authority-name prod-ca
```

If the target resources already exist, the command fails fast by default. Use `--force` to update existing resources instead:

```bash
multikubectl import prod --force
```

This workflow is useful when you want to bootstrap Multikube from an existing cluster configuration instead of manually creating CA, credential, certificate, and backend resources one by one.

## Create a backend with the CLI

```bash
multikubectl create backend prod-cluster \
  --server https://prod-api.example.internal:6443 \
  --ca-ref prod-ca \
  --auth-ref prod-token \
  --cache-ttl 30s
```

Useful flags:

- `--server` required upstream API server URL. Can be used multiple times.
- `--ca-ref` reference to a CA resource.
- `--auth-ref` reference to a credential resource.
- `--insecure-skip-tls-verify` disables upstream certificate verification.
- `--cache-ttl` duration such as `30s`, `5m`, or `1h`.
- `--label` attaches one or more metadata labels.

The CLI currently exposes only the core backend connection fields. `impersonation_config`, `probes`, `type`, and `enabled` are currently YAML or API driven settings.

## Current behavior and caveats

- `servers` must contain at least one upstream API server URL.
- `ca_ref` must point to an existing CA resource or compilation fails when verification depends on it.
- `auth_ref` must point to an existing credential resource or compilation fails.
- if the credential referenced by `auth_ref` points to a client certificate, that certificate must exist or compilation fails.
- `cache_ttl` defaults to `30s` when omitted on create or full update.
- `type` defaults to `LOAD_BALANCING_TYPE_ROUND_ROBIN` when omitted on create or full update.
- `enabled` defaults to `true` when omitted on create or full update.
- `impersonation_config` defaults to the built-in `default` profile when omitted on create or full update.
- `probes` defaults to built-in `/healthz` and `/readyz` probes when omitted on create or full update.
- client-supplied `Impersonate-*` headers are always stripped before forwarding.
- `multikubectl import` supports kubeconfig contexts that use one supported auth method at a time: token, basic auth, or client certificate auth.
- `multikubectl import` does not currently support kubeconfig `exec` plugins or legacy `auth-provider` entries.
- `multikubectl import` reads referenced files relative to the kubeconfig file, which makes it work with normal kubeconfig CA, token, and client certificate file references.

