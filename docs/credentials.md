# Credentials

<!--toc:start-->
- [Credentials](#credentials)
  - [What a credential is](#what-a-credential-is)
  - [Supported credential types](#supported-credential-types)
  - [When to use a credential](#when-to-use-a-credential)
  - [Basic examples](#basic-examples)
  - [Example scenarios](#example-scenarios)
  - [Create a credential with the CLI](#create-a-credential-with-the-cli)
  - [Current behavior and caveats](#current-behavior-and-caveats)
  - [Full Config](#full-config)
<!--toc:end-->

A credential defines how Multikube authenticates to an upstream backend.

## What a credential is

Backends do not store auth material directly. Instead, a backend uses `auth_ref` to reference a credential resource.

This keeps connection details and authentication details separate:

- the backend says where to connect
- the credential says how to authenticate

## Supported credential types

A credential can contain exactly one authentication method:

- bearer token with `token`
- basic auth with `basic.username` and `basic.password`
- client certificate authentication with `client_certificate_ref`

## When to use a credential

Create a credential when the upstream Kubernetes API server requires:

- a static bearer token
- HTTP basic authentication
- mutual TLS with a client certificate and private key

## Basic examples

Bearer token:

```yaml
version: credential/v1
meta:
  name: prod-token
config:
  name: prod-token
  token: eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.example
```

Basic auth:

```yaml
version: credential/v1
meta:
  name: legacy-basic
config:
  name: legacy-basic
  basic:
    username: api-user
    password: change-me
```

Client certificate:

```yaml
version: credential/v1
meta:
  name: mtls-client
config:
  name: mtls-client
  client_certificate_ref: upstream-client-cert
```

## Example scenarios

### Bearer token for a service account or external issuer

Use a token credential when the upstream cluster accepts bearer tokens.

```yaml
version: credential/v1
meta:
  name: staging-token
config:
  name: staging-token
  token: <staging-access-token>
```

When a backend references this credential, Multikube injects:

```text
Authorization: Bearer <token>
```

### Basic auth for legacy fronted APIs

```yaml
version: credential/v1
meta:
  name: basic-upstream
config:
  name: basic-upstream
  basic:
    username: integration-user
    password: integration-password
```

When a backend references this credential, Multikube injects a standard HTTP `Authorization: Basic ...` header.

### Client certificate for mutual TLS

```yaml
version: credential/v1
meta:
  name: prod-mtls
config:
  name: prod-mtls
  client_certificate_ref: prod-client-cert
```

Use this when the upstream server expects TLS client authentication instead of an authorization header.

## Create a credential with the CLI

Bearer token:

```bash
multikubectl create credential prod-token \
  --token '<token>'
```

Basic auth:

```bash
multikubectl create credential legacy-basic \
  --basic-username api-user \
  --basic-password change-me
```

Client certificate:

```bash
multikubectl create credential prod-mtls \
  --certificate-ref prod-client-cert
```

Useful flags:

- `--token` bearer token value
- `--basic-username` basic auth username
- `--basic-password` basic auth password
- `--certificate-ref` certificate resource used for mTLS
- `--label` attach metadata labels

## Current behavior and caveats

- exactly one auth method must be set
- if `client_certificate_ref` is used, it must point to an existing certificate resource once a backend compiles against it
- token credentials inject `Authorization: Bearer <token>`
- basic auth credentials inject `Authorization: Basic <base64(username:password)>`
- client certificate credentials do not inject headers; they attach a TLS client certificate to the backend TLS config
- credentials have a `healthy` status field, but it is not used as a compile-time filter

## Full Config

```yaml
version: credential/v1 # Required. API version for credential resources.
meta:
  name: prod-token # Required. Resource name. Must be unique for this credential resource type.
  labels: # Optional. Arbitrary labels for grouping and filtering credentials.
    env: production
    auth: upstream
  created: "2026-04-04T12:00:00Z" # Server-managed. Creation timestamp.
  updated: "2026-04-04T12:00:00Z" # Server-managed. Last update timestamp.
  generation: 1 # Server-managed. Desired-state generation.
  resource_version: 1 # Server-managed. Internal revision number.
  uid: "11111111-2222-3333-4444-555555555555" # Server-managed. Unique identifier.
config:
  name: prod-token # Optional but normally set to the same value as meta.name.
  token: <bearer-token> # Optional. Bearer token string used for Authorization header injection.
  # client_certificate_ref: prod-client-cert # Optional. Name of a certificate resource for mTLS auth.
  # basic: # Optional. Basic auth credentials.
  #   username: api-user # Required when using basic auth.
  #   password: api-password # Required when using basic auth.
status:
  healthy: true # Runtime status field. Present in the schema but not used to exclude credentials from compilation.
```

Field notes:

- `version` should be `credential/v1`
- `meta.name` is the identifier referenced by `backend.config.auth_ref`
- `config.name` is part of the schema and is usually the same as `meta.name`
- choose exactly one of `client_certificate_ref`, `token`, or `basic`
- `token` is stored as plain string data in the resource payload
- `basic.username` and `basic.password` must both be set together
- `client_certificate_ref` should reference a certificate resource that contains both certificate and private key material
