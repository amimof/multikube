# Configuring Multikube to use Dex as OIDC provider

Multikube supports OIDC providers that support [OpenID Connect Discovery](https://openid.net/specs/openid-connect-discovery-1_0.html) since it uses the `/.well-known/openid-configuration` endpoint for automatic configuration. We can use [Dex](https://github.com/dexidp/dex/) for that, which is an excellent OIDC Provider written in Go.

If you have not already created a kubeconfig for Multikube then make sure to do so first. Read [kubeconfig-example](https://github.com/amimof/multikube/blob/master/docs/examples/kubeconfig-example.md) of how to create it.

1. Follow Dex's [Getting Started](https://github.com/dexidp/dex/blob/master/Documentation/getting-started.md) guide of how to download and run Dex with the included config-dev configuration.
2. Run the example application included in Dex, also available in the getting started guide.
3. Run Multikube with Dex as OIDC provider
```
multikube \
  --oidc-issuer-url="http://localhost:5556/dex/" \
  --tls-certificate=/etc/multikube/server.pem \
  --tls-key=/etc/multikube/server-key.pem 
```

