# Installing on Kubernetes

You can run Multikube on kubernetes. Deploy manifests are in `deploy/` and uses `emptyDir` ephemeral storage. Make sure to use persistent volumes in production or control plane state will be lost when multikube pods restart.

```bash
kubectl apply -f https://raw.githubusercontent.com/amimof/multikube/refs/heads/master/deploy/k8s.yaml
```

That command will deploy a single replica multikube Deployment with default configuration and self-signed generated certificates.
