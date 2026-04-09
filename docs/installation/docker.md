# Running with Docker/Podman/Nerdctl

You can run Multikube with container engines such as Docker, Podman, Nerdctl and more. Images are published to [ghcr.io/amimof/multikube](https://github.com/amimof/multikube/pkgs/container/multikube) on new releases.

```bash
docker run -d \
  --name multikube \
  -p 5743:5743 \
  -p 8443:8443 \
  -v multikube-data:/.local/state/multikube  \
  multikube:latest
```

That command will run Multikube control plane in a Docker container with default configuration and self-signed generated certificates.
