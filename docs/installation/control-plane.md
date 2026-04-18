# Installing Multikube

# Docker

You can run Multikube with container engines such as Docker, Podman, Nerdctl and more. Images are published to [ghcr.io/amimof/multikube](https://github.com/amimof/multikube/pkgs/container/multikube) on new releases.

```bash
docker run -d \
  --name multikube \
  -p 5743:5743 \
  -p 8443:8443 \
  -v multikube-data:/.local/state/multikube  \
  multikube:latest
```

## Kubernetes

Kubernetes deployment manifests are in `deploy/` and uses `emptyDir` ephemeral storage. Make sure to use persistent volumes in production or control plane state will be lost when multikube pods restart.

```bash
kubectl apply -f https://raw.githubusercontent.com/amimof/multikube/refs/heads/master/deploy/k8s.yaml
```


## Systemd

1. Download from [GitHub Releases](https://github.com/amimof/multikube/releases)

   ```bash
   curl -LO https://github.com/amimof/multikube/releases/download/v1.0.0-beta.2/multikube-linux-amd64
   ```

   > Make sure to download the binary that matches your system configuration

2. Move it in place either manually or with `install`

   ```bash
   sudo install multikube-linux-amd64 /usr/local/bin/multikube
   ```

   or

   ```bash
   sudo mv multikube-linux-amd64 /usr/local/bin/multikube
   ```

3. Create `/etc/systemd/system/multikube.service`

   ```toml
   [Unit]
   Description=Multikube
   Documentation=https://github.com/amimof/multikube
   After=network-online.target
   Wants=network-online.target
 
   [Service]
   Type=simple
 
   #  Adjust binary path if installed elsewhere
   ExecStart=/usr/local/bin/multikube
 
   Restart=on-failure
   RestartSec=5s
 
   #  Hardening (tune as needed)
   NoNewPrivileges=true
   LimitNOFILE=65535
 
   [Install]
   WantedBy=multi-user.target
   ```

4. Reload, Enable and Start

   ```bash
   systemctl daemon-reload
   systemctl enable --now multikube.service
   ```

