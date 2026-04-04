# Getting Started

<!--toc:start-->
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Running the Control Plane](#running-the-control-plane)
  - [Downloading and Installing the CLI](#downloading-and-installing-the-cli)
  - [Configuring the CLI](#configuring-the-cli)
  - [Next Steps](#next-steps)
<!--toc:end-->

This guide walks you through running a single-instance containerized Multikube control plane and installing the multikubectl cli to manage it.

## Prerequisites

- A container runtime, [Docker](https://docker.io) is used in this guide
- Internet connection

## Running the Control Plane

1. **Running `multikube` with Docker**

   First thing you need to do is to run the control plane. The control plane exposes the management API and the router in a single binary. You'll then use `multikubectl` to manage the state of the control plane. Below command will run multikube with default configuration. Pass `--help` to see a list of available flags.

   ```bash
   docker run -d \
     --name multikube \
     -p 5743:5743 \
     -p 8443:8443 \
     -v multikube-data:/.local/state/multiUserCacheDirkube  \
     multikube:latest
   ```

## Downloading and Installing the CLI

1. **Download `multikubectl`**

   Next install the multikubectl cli binary. Easiest way of doing so is to download from [GitHub Releases](https://github.com/amimof/multikube/releases). Download the latest binary from the [release page](https://github.com/amimof/multikube/releases) for your target platform with `cURL`. Below is for Linux.

   > I'm working on publishing it on popular repositories

   ```bash
   curl -LOs https://github.com/amimof/multikube/releases/latest/download/multikube-linux-amd64
   ```

2. **Install `multikubectl`**

   Place the downloaded binary in your `$PATH`. Either use the `install` command or move the binary manually

   ```bash
   sudo install multikube-darwin-amd64 /usr/local/bin/multikube
   ```

   or

   ```bash
   sudo mv multikube-darwin-amd64 /usr/local/bin/multikube
   ```

3. Verify installation

   Following command

   ```bash
   multikubectl version
   ```

   Should give you something like this:

   ```
   Version: v1.0.0-beta.1-19-gaeb75b2-dirty
   Commit: aeb75b2c896a209bf58cf87f00ee66e6354fdbb9
   ```

## Configuring the CLI

Multikubectl reads it's configuration from a yaml file located at `~/.multikube/multikube.yaml`. That file contains connection details and credentials to one or more multikube control planes. We simply use the `create-server` command to *connect* to the multikube instance you ran earlier with Docker.

1. **Adding servers**

   ```bash
   multikubectl config init
   multikubectl config create-server my-cluster --address localhost:5743 --tls --insecure
   ```

   > The `init` command creates an empty config file at `~/.multikube/multikube.yaml` and is only necessary if you've never used multikubectl before.

2. **List backends**

  Verify that multikubectl can communicate with the control plane by listing backends. You should see an empy list because we haven't created any backends yet.

   ```bash
   multikubectl get backends
   ```

🎉 You have successfully setup multikube! The next step is to add some backends and routes. Continue reading the docs to learn how.

## Next Steps

- For a more detailed installation guide see the [Installation Guide](/docs/installation/README.md).
- To learn more read the [Documentation](/docs/README.md)
