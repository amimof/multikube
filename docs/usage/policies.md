# Policies

<!--toc:start-->
- [Policies](#policies)
  - [What a policy is](#what-a-policy-is)
  - [How policy evaluation works](#how-policy-evaluation-works)
  - [When to use a policy](#when-to-use-a-policy)
  - [Basic example](#basic-example)
  - [Example scenarios](#example-scenarios)
  - [Create a policy with the CLI](#create-a-policy-with-the-cli)
  - [Current behavior and caveats](#current-behavior-and-caveats)
  - [Full Config](#full-config)
<!--toc:end-->

A policy defines authorization rules for requests that have already matched a route and backend.

## What a policy is

Policies are global. They are not attached directly to a backend or route.

Instead, Multikube evaluates all policies against:

- the authenticated principal
- the selected backend
- the parsed Kubernetes request

Each policy contains zero or more rules. Each rule can match on:

- subjects such as users, groups, service accounts, and claims
- clusters by name or label
- Kubernetes resources such as api group, resource, namespace, name, and subresource
- actions such as `GET`, `LIST`, `WATCH`, `CREATE`, and `DELETE`

## How policy evaluation works

Multikube uses deny-first evaluation:

1. if any matching rule has effect `EFFECT_DENY`, the request is denied
2. otherwise, if any matching rule has effect `EFFECT_ALLOW`, the request is allowed
3. otherwise, the request is denied

Wildcard behavior is important:

- no `subjects` means any subject
- no `clusters` means any backend
- no `resources` means any Kubernetes resource
- no `actions` means any action

## When to use a policy

Create policies when you want to:

- allow read-only access for a team
- limit users to a specific namespace or resource kind
- deny dangerous operations on production clusters
- separate access between clusters selected by route rules

## Basic example

```yaml
version: policy/v1
meta:
  name: platform-readonly
config:
  name: platform-readonly
  rules:
    - effect: EFFECT_ALLOW
      subjects:
        - groups:
            - platform
      resources:
        - resource: pods
      actions:
        - ACTION_GET
        - ACTION_LIST
        - ACTION_WATCH
```

This allows members of the `platform` group to read pods.

## Example scenarios

### Read-only access to one namespace

```yaml
version: policy/v1
meta:
  name: team-a-readonly
config:
  name: team-a-readonly
  rules:
    - effect: EFFECT_ALLOW
      subjects:
        - groups:
            - team-a
      resources:
        - resource: pods
          namespaces:
            - team-a
        - resource: deployments
          namespaces:
            - team-a
      actions:
        - ACTION_GET
        - ACTION_LIST
        - ACTION_WATCH
```

Use this when a team should observe resources in its own namespace but not mutate them.

### Deny deletes on production clusters

```yaml
version: policy/v1
meta:
  name: deny-prod-delete
config:
  name: deny-prod-delete
  rules:
    - effect: EFFECT_DENY
      clusters:
        - names:
            - prod-cluster
      actions:
        - ACTION_DELETE
        - ACTION_DELETECOLLECTION
```

Because deny rules win, this blocks delete operations even if another policy also allows them.

### Allow one user to access one cluster

```yaml
version: policy/v1
meta:
  name: alice-dev-access
config:
  name: alice-dev-access
  rules:
    - effect: EFFECT_ALLOW
      subjects:
        - users:
            - alice@example.com
      clusters:
        - names:
            - dev-cluster
```

Because resources and actions are omitted, this rule behaves as a wildcard for those dimensions.

### Match a custom claim

```yaml
version: policy/v1
meta:
  name: tenant-acme
config:
  name: tenant-acme
  rules:
    - effect: EFFECT_ALLOW
      subjects:
        - claims:
            - name: tenant
              value: acme
      clusters:
        - names:
            - acme-cluster
```

Use this when identities carry tenant or ownership information in JWT claims.

## Create a policy with the CLI

The current CLI creates the policy resource shell but does not expose rule-building flags.

```bash
multikubectl create policy platform-readonly
```

In practice, richer policies are easiest to create or update through the API payload format shown in the YAML examples.

## Current behavior and caveats

- policies are global and are evaluated together; there is no `policy_ref` on routes or backends
- if no policy allows a request, the request is denied by default
- a matching deny rule always wins over matching allow rules
- cluster label matching currently succeeds if any one label matches, not only when all labels match
- `conditions` are defined in the schema but are not currently evaluated
- `resource.label_selector` is defined in the schema but is not currently evaluated
- `ACTION_EXEC`, `ACTION_PORTFORWARD`, and `ACTION_PROXY` are defined but do not currently match requests in the evaluator
- `ACTION_LOGS` maps to `get`
- `ACTION_LIST` matches both `list` and `get`

## Full Config

```yaml
version: policy/v1 # Required. API version for policy resources.
meta:
  name: platform-readonly # Required. Resource name. Must be unique for this policy resource type.
  labels: # Optional. Arbitrary labels for grouping and filtering policies.
    env: production
    security: access
  created: "2026-04-04T12:00:00Z" # Server-managed. Creation timestamp.
  updated: "2026-04-04T12:00:00Z" # Server-managed. Last update timestamp.
  generation: 1 # Server-managed. Desired-state generation.
  resource_version: 1 # Server-managed. Internal revision number.
  uid: "11111111-2222-3333-4444-555555555555" # Server-managed. Unique identifier.
config:
  name: platform-readonly # Optional but normally set to the same value as meta.name.
  rules:
    - effect: EFFECT_ALLOW # Required for useful rules. Valid values: EFFECT_ALLOW, EFFECT_DENY.
      subjects: # Optional. Empty means any authenticated principal.
        - users: # Optional. Exact user identifiers.
            - alice@example.com
          groups: # Optional. Group names.
            - platform
          service_accounts: # Optional. Service account identifiers.
            - system:serviceaccount:ops:deployer
          claims: # Optional. Arbitrary principal claim matches.
            - name: tenant
              value: acme
      clusters: # Optional. Empty means any backend cluster.
        - names:
            - prod-cluster
          labels:
            env: production
            team: platform
      resources: # Optional. Empty means any Kubernetes resource.
        - api_group: apps # Optional. Kubernetes API group, for example apps or batch.
          resource: deployments # Optional. Resource type, for example pods or deployments.
          sub_resource: status # Optional. Subresource such as status, scale, or log.
          namespaces: # Optional. Namespace allow-list.
            - platform
          names: # Optional. Resource name allow-list.
            - api-server
          label_selector: # Defined in schema, but not currently enforced by the evaluator.
            match_labels:
              app: api-server
      actions: # Optional. Empty means any action.
        - ACTION_GET
        - ACTION_LIST
        - ACTION_WATCH
        - ACTION_CREATE
        - ACTION_UPDATE
        - ACTION_PATCH
        - ACTION_DELETE
        - ACTION_DELETECOLLECTION
        - ACTION_PROXY
        - ACTION_EXEC
        - ACTION_LOGS
        - ACTION_PORTFORWARD
      conditions: # Defined in schema, but not currently enforced by the evaluator.
        - type: CONDITION_TYPE_SOURCEIP # Valid types also include TIMEWINDOW, WEEKDAY, and CLAIM.
status: {} # Present in the schema. Currently empty.
```

Field notes:

- `version` should be `policy/v1`
- `meta.name` is the stable identifier for the policy resource
- `config.name` is part of the schema and is usually the same as `meta.name`
- `rules` is optional, but an empty policy grants nothing by itself
- `subjects`, `clusters`, `resources`, and `actions` each behave as wildcards when omitted
- `effect` determines whether a matching rule allows or denies the request
- selectors inside each repeated list are OR-style at the top level: if any selector entry matches, that dimension matches
