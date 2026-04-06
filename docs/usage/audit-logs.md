# Audit Logs

<!--toc:start-->
- [Audit Logs](#audit-logs)
  - [Where logs are written](#where-logs-are-written)
  - [How logging works](#how-logging-works)
  - [Log format](#log-format)
  - [Example event](#example-event)
  - [Fields you will see](#fields-you-will-see)
  - [Operational notes](#operational-notes)
  - [Current behavior and caveats](#current-behavior-and-caveats)
<!--toc:end-->

Audit logs record requests that pass through Multikube so operators can review who accessed the proxy, what path was requested, and how the request completed. The audit log is a structured event stream produced by the proxy. Each event captures:

- when the request happened
- who made the request
- which HTTP path and method were used
- which Kubernetes resource the request targeted
- whether the request was allowed or denied
- how the request completed

## Where logs are written

When `multikube` starts, it creates a file sink for audit events at:

```text
<data-path>/audit.json
```

The `data-path` value comes from the `--data-path` flag used when starting the server.

Example:

```bash
multikube --data-path /var/lib/multikube
```

With that configuration, audit events are appended to:

```text
/var/lib/multikube/audit.json
```

## How logging works

At startup, Multikube creates:

- a file sink that appends events to disk
- an asynchronous publisher that batches events before writing them
- HTTP audit middleware that creates an audit event for each request

The publisher is buffered and flushes events in batches, which reduces write overhead compared to writing each request synchronously.

## Log format

Audit events are written as newline-delimited JSON.

- one JSON object per line
- suitable for `jq`, log shippers, and streaming parsers
- appended to the same file over time

This means you can inspect the file with tools such as:

```bash
jq . /var/lib/multikube/audit.json
```

Or stream new events as they arrive:

```bash
tail -f /var/lib/multikube/audit.json
```

## Example event

An audit record looks like this:

```json
{"ts":"2026-04-05T11:33:12.145112Z","method":"GET","path":"/api/v1/namespaces/default/pods","source_ip":"10.0.0.25","user_agent":"kubectl/v1.31.0","status_code":200,"duration_ms":18}
```

Depending on request flow and future enrichment, additional fields may also be present.

## Fields you will see

The audit event schema in `pkg/audit/audit.go` includes these fields:

- identity: `request_id`, `subject`, `username`, `groups`, `issuer`
- target selection: `cluster`, `backend`, `route`
- HTTP request: `method`, `path`, `source_ip`, `user_agent`
- Kubernetes request: `k8s_verb`, `api_group`, `resource`, `namespace`, `name`, `subresource`
- outcome: `allowed`, `status_code`, `duration_ms`, `error`
- config context: `policy_ids`, `config_version`

Typical meanings:

- `ts` is the event timestamp
- `method` is the incoming HTTP method such as `GET` or `POST`
- `path` is the original request path seen by the proxy
- `source_ip` is derived from the client TCP remote address
- `user_agent` is copied from the incoming request
- `status_code` is the response status returned to the client
- `duration_ms` is total request time in milliseconds

## Operational notes

- the log file is append-only from Multikube's point of view
- records are buffered and written asynchronously
- the file format is machine-friendly rather than human-formatted
- external log rotation should be used if you expect the file to grow continuously

Because events are written locally to disk, a common production setup is to let a system service, sidecar, or log agent ship the file to a centralized logging backend.

## Current behavior and caveats

The audit log feature is present and usable, but it is still evolving.

- the current sink writes only to a local file; there is no built-in remote sink yet
- event delivery is asynchronous, so a process crash can lose buffered events that have not been flushed yet
- some audit schema fields exist for future enrichment and may be empty depending on request path and middleware order
- Kubernetes-specific request fields such as `api_group`, `resource`, and `namespace` are intended to be recorded, but are not currently populated consistently in all flows
- identity and policy-related fields may also be incomplete until more request context is attached before audit publication

If you need stronger guarantees or additional destinations, treat the current implementation as a foundation and plan for external collection, retention, and rotation around `audit.json`.
