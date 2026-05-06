# Authentication

<!--toc:start-->
- [Authentication](#authentication)
  - [How authentication works](#how-authentication-works)
  - [Login](#login)
    - [With multikubectl](#with-multikubectl)
    - [With grpcurl](#with-grpcurl)
    - [With curl](#with-curl)
  - [Users](#users)
  - [Roles](#roles)
<!--toc:end-->

Authentication in Multikube is about signing in to the control plane so you can use `multikubectl`, the API, and the web console. For most users, the normal way to authenticate is to log in with `multikubectl` and let the CLI store the session for later requests.

API Reference:
- [auth/v1](https://github.com/amimof/multikube/blob/master/api/auth/v1/auth.proto)
- [user/v1](https://github.com/amimof/multikube/blob/master/api/user/v1/user.proto)

## How authentication works

Multikube authenticates users through `AuthService` against its built-in user store. A user signs in with a username and password. When the login succeeds, Multikube returns an `access_token` and a `refresh_token`. The access token is sent on later requests as `Authorization: Bearer <access_token>`, and the refresh token is used to obtain a new access token when the old one expires.

This page only covers how a user authenticates to Multikube itself. Authentication from Multikube to upstream clusters and backends is a separate concern and is documented in [Credentials](./credentials.md).

## Login

The first time Multikube starts, a default administrator account is created with the credentials `admin/admin`. You can use that account to sign in for the first time, create additional users, and manage the rest of the control plane.

The recommended way to log in is with `multikubectl`. The CLI sends your username and password to `AuthService`, receives an access token and refresh token, and stores the session in `~/.config/multikubectl.yaml`. On later requests, the CLI uses the stored tokens to authenticate to the control plane.

### With multikubectl

If you already have a current server configured in your CLI config, logging in is as simple as running:

```bash
multikubectl login
```

If this is your first time using `multikubectl`, you can create the config file and a server entry as part of the login command itself. The following example creates a server entry named `dev`, points it at the control plane, and stores the authenticated session in one step.

```bash
multikubectl login dev --address localhost:5743 --insecure
```

When you run `multikubectl login`, credentials are resolved from the command-line flags first, then from the environment variables `MULTIKUBECTL_USERNAME` and `MULTIKUBECTL_PASSWORD`, and finally from an interactive prompt if nothing else is provided. This makes it easy to use the same command both interactively and in scripts.

You can provide `--ca`, `--certificate`, and `--key` to trust a custom certificate authority or to use mutual TLS. The `--insecure` flag skips TLS verification and is mainly useful for local development.

### With grpcurl

The gRPC management API listens on port `5743` by default. If you want to log in without the CLI, you can call `AuthService/Login` directly with `grpcurl`.

```bash
grpcurl -insecure \
  -d '{"username": "admin", "password": "admin"}' \
  localhost:5743 auth.v1.AuthService/Login
```

That response includes both `access_token` and `refresh_token`. When the access token expires, you can exchange the refresh token for a new pair of tokens.

```bash
grpcurl -insecure \
  -d '{"refresh_token": "<refresh-token>"}' \
  localhost:5743 auth.v1.AuthService/Refresh
```

### With curl

The HTTP API is exposed through `grpc-gateway` and listens on port `6443` by default. You can log in with `curl`, or with any other HTTP client that can send JSON.

```bash
curl -k \
  -H 'Content-Type: application/json' \
  -X POST \
  -d '{"username": "admin", "password": "admin"}' \
  https://localhost:6443/api/v1/auth/login
```

To refresh an expired access token, call the refresh endpoint with the refresh token you received at login.

```bash
curl -k \
  -H 'Content-Type: application/json' \
  -X POST \
  -d '{"refresh_token": "<refresh-token>"}' \
  https://localhost:6443/api/v1/auth/refresh
```

## Users

A user is the identity that signs in to the Multikube control plane. Users are stored by Multikube itself, and each user carries the roles that determine what the user is allowed to do after authentication. A user account can also be disabled, in which case it can no longer log in or refresh its session.

This page focuses on authentication, not user lifecycle management. For the user schema and API surface, see the `user/v1` API reference linked above.

## Roles

Multikube includes three built-in roles: `admin`, `viewer`, and `client`. The `admin` role has full access to the control-plane API. The `viewer` role has read-only access. The `client` role is intended for routed or proxy-facing scenarios and does not grant normal control-plane management access.

When a user logs in, the user's roles are embedded in the issued token and are used by the control plane to authorize later requests.
