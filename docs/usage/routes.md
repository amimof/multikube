# Routes

<!--toc:start-->
- [Routes](#routes)
  - [What a route is](#what-a-route-is)
  - [How route matching works](#how-route-matching-works)
  - [When to use a route](#when-to-use-a-route)
  - [Basic example](#basic-example)
  - [Example scenarios](#example-scenarios)
  - [Create a route with the CLI](#create-a-route-with-the-cli)
  - [Current behavior and caveats](#current-behavior-and-caveats)
  - [Full Config](#full-config)
<!--toc:end-->

A route decides which backend should handle an incoming request.

## What a route is

Each route contains:

- one matcher in `config.match`
- one backend reference in `config.backend_ref`

When a request matches that rule, Multikube forwards it to the referenced backend.

## How route matching works

Routes are evaluated in this runtime order:

1. `path`
2. `path_prefix`
3. `header`
4. `jwt`
5. `sni`

That order matters. A request that matches both a path-based route and a header-based route will use the path-based route first.

`config.match` supports exactly one matcher:

- `path`
- `path_prefix`
- `header`
- `jwt`
- `sni`

## When to use a route

Create routes when you want to:

- send different request paths to different clusters
- choose a backend based on a tenant header
- route requests from users with specific JWT claims
- dispatch TLS connections by SNI
- split traffic between clusters with clear and explicit rules

## Basic example

```yaml
version: route/v1
meta:
  name: prod-api
config:
  name: prod-api
  match:
    path_prefix: /prod
  backend_ref: prod-cluster
```

This route sends any request whose URL path starts with `/prod` to the backend named `prod-cluster`.

## Example scenarios

### Route by exact path pattern

```yaml
version: route/v1
meta:
  name: metrics-route
config:
  name: metrics-route
  match:
    path: /api/v1/nodes
  backend_ref: prod-cluster
```

Use this when one specific path should always go to one backend.

### Route by path prefix

```yaml
version: route/v1
meta:
  name: dev-prefix
config:
  name: dev-prefix
  match:
    path_prefix: /clusters/dev
  backend_ref: dev-cluster
```

Use this when one group of URLs should map to the same backend.

### Route by header

```yaml
version: route/v1
meta:
  name: tenant-acme
config:
  name: tenant-acme
  match:
    header:
      name: X-Tenant
      value: acme
  backend_ref: acme-cluster
```

Use this when the caller or an upstream proxy already sets a stable tenant header.

### Route by JWT claim

```yaml
version: route/v1
meta:
  name: platform-team
config:
  name: platform-team
  match:
    jwt:
      claim: team
      value: platform
  backend_ref: platform-cluster
```

Use this when authenticated users should be routed according to identity information in their JWT.

### Route by SNI

```yaml
version: route/v1
meta:
  name: cluster-sni
config:
  name: cluster-sni
  match:
    sni: api.example.com
  backend_ref: external-cluster
```

Use this when clients connect with TLS and the target cluster should be selected from the requested server name.

## Create a route with the CLI

Path prefix example:

```bash
multikubectl create route prod-api \
  --backend-ref prod-cluster \
  --path-prefix /prod
```

Header example:

```bash
multikubectl create route tenant-acme \
  --backend-ref acme-cluster \
  --header-name X-Tenant \
  --header-value acme
```

JWT example:

```bash
multikubectl create route platform-team \
  --backend-ref platform-cluster \
  --jwt-claim team \
  --jwt-value platform
```

Useful flags:

- `--backend-ref` target backend name
- `--path` path matcher
- `--path-prefix` prefix matcher
- `--sni` SNI matcher
- `--header-name` and `--header-value` header matcher pair
- `--jwt-claim` and `--jwt-value` JWT matcher pair
- `--label` attach metadata labels

## Current behavior and caveats

- `config.match` is required; routes without a matcher become `INVALID`
- `backend_ref` must point to an existing compiled backend or the route becomes `INVALID`
- duplicate matchers conflict; both conflicting routes are marked `CONFLICT` and removed from the active runtime
- `path_prefix` routes are sorted by longest prefix first, so more specific prefixes win
- `path` matching uses Go `path.Match`, which means it behaves like glob matching rather than strict literal comparison
- JWT matching only works when JWT claims are present in the request context
- SNI matching only works when SNI is present in the request context

Routes that compile successfully are marked `READY`.

## Full Config

```yaml
version: route/v1 # Required. API version for route resources.
meta:
  name: prod-api # Required. Resource name. Must be unique for this route resource type.
  labels: # Optional. Arbitrary labels for grouping and filtering routes.
    env: production
    traffic: primary
  created: "2026-04-04T12:00:00Z" # Server-managed. Creation timestamp.
  updated: "2026-04-04T12:00:00Z" # Server-managed. Last update timestamp.
  generation: 1 # Server-managed. Desired-state generation.
  resource_version: 1 # Server-managed. Internal revision number.
  uid: "11111111-2222-3333-4444-555555555555" # Server-managed. Unique identifier.
config:
  name: prod-api # Optional but normally set to the same value as meta.name.
  match: # Required. Exactly one matcher may be set.
    sni: api.example.com # Optional matcher. Match by TLS server name.
    # header: # Optional matcher. Match by HTTP header name and value.
    #   name: X-Tenant
    #   value: acme
    # path: /api/v1/nodes # Optional matcher. Uses Go path.Match glob semantics.
    # path_prefix: /clusters/prod # Optional matcher. Prefix match.
    # jwt: # Optional matcher. Match JWT claim name and exact string value.
    #   claim: team
    #   value: platform
  backend_ref: prod-cluster # Required in practice. Name of the backend that should receive matching requests.
status:
  phase: READY # Server-managed. Common values are READY, INVALID, and CONFLICT.
  reason: "" # Server-managed. Human-readable compile reason when not READY.
  last_transition_time: "2026-04-04T12:00:00Z" # Server-managed. Last phase change timestamp.
```

Field notes:

- `version` should be `route/v1`
- `meta.name` is the identifier used when listing, updating, and deleting the route
- `config.name` is part of the schema and is typically the same as `meta.name`
- `match` is a one-of field; choose only one matcher per route
- `backend_ref` should be the name of a backend resource, not a URL
- `status.phase` reflects compile outcome rather than traffic volume or request health
