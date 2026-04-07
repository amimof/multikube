# E2E Scaffolding

This directory contains a small `kind`-based end-to-end environment for local Multikube testing.

The scaffold is intentionally lightweight:

- one management cluster running `multikube`
- two workload clusters imported as backend targets
- shell scripts for setup, deployment, smoke testing, and teardown

It is meant to be a starting point for adding your own e2e scenarios later.

## Prerequisites

You need these tools available locally:

- `docker`
- `kind`
- `kubectl`
- `go`
- `curl`

## Quick Start

Run the full flow:

```bash
make e2e
```

Or step through it manually:

```bash
./e2e/setup.sh
./e2e/deploy-multikube.sh
./e2e/tests/smoke.sh
./e2e/teardown.sh
```

## What The Smoke Test Does

The included smoke test performs a basic happy path:

- creates three local `kind` clusters
- deploys `multikube` into the management cluster
- imports the two workload clusters with `multikubectl import`
- creates one header-based route per backend
- verifies proxied requests reach the expected backend cluster

To distinguish the backend clusters, the test creates unique namespaces in each cluster and then queries those namespaces through the Multikube proxy.

## Artifacts

Temporary files are written to `e2e/tmp/`:

- local kubeconfigs
- built `multikubectl` binary
- port-forward logs
- captured smoke test responses

Keep the environment around for manual debugging:

```bash
E2E_KEEP=1 make e2e
```

## Extending It

Good next additions:

- more routes and policy scenarios
- round-robin backend pool validation
- backend health and failover checks
- JWT and header matching cases
- CI integration
