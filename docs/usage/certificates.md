# Certificates

<!--toc:start-->
- [Certificates](#certificates)
  - [What this document covers](#what-this-document-covers)
  - [Certificates vs CAs](#certificates-vs-cas)
  - [How these resources are used](#how-these-resources-are-used)
  - [When to use them](#when-to-use-them)
  - [Basic examples](#basic-examples)
  - [Example scenarios](#example-scenarios)
  - [Create certificates and CAs with the CLI](#create-certificates-and-cas-with-the-cli)
  - [Current behavior and caveats](#current-behavior-and-caveats)
  - [Full Config: Certificate](#full-config-certificate)
  - [Full Config: Certificate Authority](#full-config-certificate-authority)
<!--toc:end-->

This page covers both certificate resources and certificate authority resources.

## What this document covers

Multikube uses two related TLS resource types:

- `certificate/v1` for client certificates and private keys
- `ca/v1` for certificate authorities used to verify upstream servers

## Certificates vs CAs

Use a certificate resource when Multikube must present a client certificate to the upstream backend.

Use a CA resource when Multikube must verify the backend server certificate.

In other words:

- certificate = what Multikube presents
- CA = what Multikube trusts

## How these resources are used

The usual chain looks like this:

1. a backend references a CA with `ca_ref`
2. a backend references a credential with `auth_ref`
3. that credential may reference a certificate with `client_certificate_ref`

At compile time:

- certificate resources become `tls.Certificate` objects
- CA resources become `x509.CertPool` objects
- backends consume those objects when building outbound TLS config

## When to use them

Create these resources when:

- the upstream API server uses a private or self-signed CA
- the upstream API server requires mutual TLS
- you want TLS verification without embedding PEM data directly into backend resources

## Basic examples

Certificate resource:

```yaml
version: certificate/v1
meta:
  name: upstream-client-cert
config:
  name: upstream-client-cert
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...client certificate PEM...
    -----END CERTIFICATE-----
  key: |
    -----BEGIN PRIVATE KEY-----
    ...private key PEM...
    -----END PRIVATE KEY-----
```

CA resource:

```yaml
version: ca/v1
meta:
  name: upstream-ca
config:
  name: upstream-ca
  certificate_data: |
    -----BEGIN CERTIFICATE-----
    ...CA certificate PEM...
    -----END CERTIFICATE-----
```

## Example scenarios

### Verify a private upstream API server certificate

Create a CA resource and reference it from the backend.

```yaml
version: ca/v1
meta:
  name: prod-ca
config:
  name: prod-ca
  certificate_data: |
    -----BEGIN CERTIFICATE-----
    ...CA PEM...
    -----END CERTIFICATE-----
```

```yaml
version: backend/v1
meta:
  name: prod-cluster
config:
  name: prod-cluster
  server: https://prod-api.example.internal:6443
  ca_ref: prod-ca
```

Use this when the upstream server certificate is not signed by a public CA trusted by the system.

### Connect with mutual TLS

Create a certificate resource, reference it from a credential, and reference that credential from a backend.

```yaml
version: certificate/v1
meta:
  name: prod-client-cert
config:
  name: prod-client-cert
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...client certificate PEM...
    -----END CERTIFICATE-----
  key: |
    -----BEGIN PRIVATE KEY-----
    ...private key PEM...
    -----END PRIVATE KEY-----
```

```yaml
version: credential/v1
meta:
  name: prod-mtls
config:
  name: prod-mtls
  client_certificate_ref: prod-client-cert
```

```yaml
version: backend/v1
meta:
  name: prod-cluster
config:
  name: prod-cluster
  server: https://prod-api.example.internal:6443
  ca_ref: prod-ca
  auth_ref: prod-mtls
```

This is the standard setup when the upstream cluster requires both:

- server verification with a trusted CA
- client authentication with a certificate

### Keep TLS material separate from backends

You can reuse one CA or client certificate across multiple backends by referencing the same resource name from several backends or credentials.

This is useful when:

- several backends share the same issuing CA
- one client certificate is valid for several upstream environments

## Create certificates and CAs with the CLI

Certificate:

```bash
multikubectl create certificate prod-client-cert \
  --certificate "$(cat client.crt)" \
  --key "$(cat client.key)"
```

CA:

```bash
multikubectl create ca prod-ca \
  --certificate "$(cat ca.crt)"
```

Useful certificate flags:

- `--certificate` inline PEM certificate string
- `--certificate-data` additional certificate field
- `--key` inline PEM private key string
- `--key-data` additional key field
- `--label` attach metadata labels

Useful CA flags:

- `--certificate` certificate field
- `--certificate-data` certificate data field
- `--label` attach metadata labels

## Current behavior and caveats

- certificate resources require one certificate field and one key field
- CA resources require one certificate field
- for certificate resources, both `certificate` and `certificate_data` are treated as inline PEM content by the compiler
- for certificate resources, both `key` and `key_data` are treated as inline PEM content by the compiler
- for CA resources, `certificate_data` is treated as inline PEM content
- for CA resources, `certificate` is treated as a reference to a certificate resource, not as inline PEM content
- that means the CA resource behaves differently from the certificate resource even though the field names look similar
- invalid PEM data or missing referenced resources cause compilation to fail

Because of the current implementation, `certificate_data` and `key_data` should be documented as content fields, not file path fields.

## Full Config: Certificate

```yaml
version: certificate/v1 # Required. API version for certificate resources.
meta:
  name: upstream-client-cert # Required. Resource name. Must be unique for this certificate resource type.
  labels: # Optional. Arbitrary labels for grouping and filtering certificates.
    env: production
    tls: client
  created: "2026-04-04T12:00:00Z" # Server-managed. Creation timestamp.
  updated: "2026-04-04T12:00:00Z" # Server-managed. Last update timestamp.
  generation: 1 # Server-managed. Desired-state generation.
  resource_version: 1 # Server-managed. Internal revision number.
  uid: "11111111-2222-3333-4444-555555555555" # Server-managed. Unique identifier.
config:
  name: upstream-client-cert # Optional but normally set to the same value as meta.name.
  certificate: |
    -----BEGIN CERTIFICATE-----
    ...PEM certificate content...
    -----END CERTIFICATE-----
  # certificate_data: | # Alternate certificate field. Current compiler also treats this as inline PEM content.
  #   -----BEGIN CERTIFICATE-----
  #   ...alternate PEM certificate content field...
  #   -----END CERTIFICATE-----
  key: |
    -----BEGIN PRIVATE KEY-----
    ...PEM private key content...
    -----END PRIVATE KEY-----
  # key_data: | # Alternate key field. Current compiler also treats this as inline PEM content.
  #   -----BEGIN PRIVATE KEY-----
  #   ...alternate PEM private key content field...
  #   -----END PRIVATE KEY-----
status: {} # Present in the schema. Currently empty.
```

Field notes:

- `version` should be `certificate/v1`
- `meta.name` is the identifier referenced by `credential.config.client_certificate_ref`
- `config.name` is part of the schema and is usually the same as `meta.name`
- set exactly one of `certificate` or `certificate_data`
- set exactly one of `key` or `key_data`
- all four payload fields are interpreted as PEM content by the current compiler

## Full Config: Certificate Authority

```yaml
version: ca/v1 # Required. API version for CA resources.
meta:
  name: upstream-ca # Required. Resource name. Must be unique for this CA resource type.
  labels: # Optional. Arbitrary labels for grouping and filtering CAs.
    env: production
    tls: trust
  created: "2026-04-04T12:00:00Z" # Server-managed. Creation timestamp.
  updated: "2026-04-04T12:00:00Z" # Server-managed. Last update timestamp.
  generation: 1 # Server-managed. Desired-state generation.
  resource_version: 1 # Server-managed. Internal revision number.
  uid: "11111111-2222-3333-4444-555555555555" # Server-managed. Unique identifier.
config:
  name: upstream-ca # Optional but normally set to the same value as meta.name.
  certificate: upstream-ca-cert # Current compiler behavior: treated as a reference to a certificate resource.
  # certificate_data: | # Alternate field. Treated as inline CA PEM content.
  #   -----BEGIN CERTIFICATE-----
  #   ...CA PEM content...
  #   -----END CERTIFICATE-----
status: {} # Present in the schema. Currently empty.
```

Field notes:

- `version` should be `ca/v1`
- `meta.name` is the identifier referenced by `backend.config.ca_ref`
- `config.name` is part of the schema and is usually the same as `meta.name`
- set exactly one of `certificate` or `certificate_data`
- `certificate_data` is interpreted as inline CA PEM content
- `certificate` is currently interpreted as the name of a certificate resource whose `config.certificate` field contains inline PEM data
