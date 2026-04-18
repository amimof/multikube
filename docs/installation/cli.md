# Installing Multikubectl


1. **Download `multikubectl`**

   Next install the multikubectl cli binary. Easiest way of doing so is to download from [GitHub Releases](https://github.com/amimof/multikube/releases). Download the latest binary from the [release page](https://github.com/amimof/multikube/releases) for your target platform with `cURL`. Below is for Linux.

   > I'm working on publishing it on popular repositories

   ```bash
   curl -LOs https://github.com/amimof/multikube/releases/latest/download/multikubectl-linux-amd64
   ```

2. **Install `multikubectl`**

   Place the downloaded binary in your `$PATH`. Either use the `install` command or move the binary manually

   ```bash
   sudo install multikubectl-darwin-amd64 /usr/local/bin/multikubectl
   ```

   or

   ```bash
   sudo mv multikubectl-darwin-amd64 /usr/local/bin/multikubectl
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

