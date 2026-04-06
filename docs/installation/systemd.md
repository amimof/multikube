# Running with Systemd

You can run Multikube with `systemd` using below unit file. Before doing so however you must download the binary and install it in your systems `PATH`.

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

That command will run Multikube control plane with default configuration and self-signed generated certificates.
