# Authentication

<!--toc:start-->
- [Authentication](#authentication)
  - [How authentication works](#how-authentication-works)
  - [Login](#login)
    - [With multikubectl](#with-multikubectl)
    - [With grpcul](#with-grpcul)
    - [With curl](#with-curl)
  - [Roles](#roles)
<!--toc:end-->

API Reference: 
- [auth/v1](https://github.com/amimof/multikube/blob/master/api/auth/v1/auth.proto)
- [user/v1](https://github.com/amimof/multikube/blob/master/api/user/v1/user.proto)

## How authentication works

Multikube authenticates users via the `AuthService`, looking up users from a builtin user registry. Currently the authentication method is using username and password, all though support for alternative mechanisms and federations is in the roadmap. After a successful `/Login`, the AuthService returns a short-lived `access_token` (15m) and a longer-lived `refresh_token` (24h) that can be used in subsequent requests to other services that requires authenticated users.

## Login

The first time Multikube starts, a default admin user is created automatically. The username and password is simply `admin/admin`. This account can be used to create additional users and to issue tokens used to communicate with upstream backends through the proxy.

### With multikubectl

Use the `multikubectl login` command to login to the controll plane with your credentials. 

```bash
multikubectl login
```

It's possible to bootstrap a multikubectl config if one does not exist yet. For example if you're setting up multikube/multikubectl for the first time. Following command will create `~/.config/multikubectl.yaml` if it doesn't exist, create a server-entry called `dev` and login in one sweep.

```bash
multikubectl login dev --address localhost:5743 --insecure
```

### With grpcul

Login with `grpcurl`

```bash
grpcurl -insecure  -d '{"username": "admin", "password": "admin"}' localhost:5743 auth.v1.AuthService/Login
```

### With curl

Login with curl, or any other http client

```bash
curl -X POST -k -d '{"username": "admin", "password": "admin"}' https://localhost:6443/api/v1/auth/login
```

## Users

## Roles

Multikube comes with a few builtin roles nameley `admin`, `viewer` and `client`. The admin and viewer roles are obvious. The `client` role is used only for routing and authenticating requests going through the proxy to backends. 
