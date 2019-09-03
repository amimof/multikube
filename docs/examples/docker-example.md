# Docker

You may use the tiny official Docker images to run Multikube. Following will run Multikube (non-tls) in a Docker container.

```
docker run -v $PWD:/etc/multikube/kubeconfig \
  amimof/multikube:latest \
  --scheme http
```