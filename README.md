# restart-compose

A blunt force, inflexible webhook listener for automatically updating Git repositories and restarting Docker Compose stacks. It utilizes Tailscale Funnel to expose a webhook endpoint to the internet. Once configured in your GitHub repository settings, `push` events received will trigger a `git pull` and a restart of configured Compose stacks, if modified.

This project serves as the v0 to a more elegant GitOps solution that I am still working on, `composer`.

## Configuration

Create a `.restart-compose.yaml` file in your working directory with a list of `stacks`:

```yaml
stacks:
  - "nginx-one/compose.yaml"
  - "nginx-two/compose.yaml"
```

The paths are relative to the directory containing the configuration file. You will also need some environment variables:

```env
TS_AUTHKEY=optionally-your-authkey
GH_SECRET=your-webhook-secret
```

Finally, you need to run the `restart-compose` container itself, which can be done using Docker Compose:

```yaml
name: "restart-compose"
services:
  watcher:
    image: "ghcr.io/hugginsio/restart-compose:v0"
    restart: "unless-stopped"
    logging:
      driver: "local"
      options:
        max-file: "3"
        max-size: "5m"
    env_file: [".env"]
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock" # required for container management
      - "restart-compose-tsnet:/tmp/restart-compose" # required for tailnet persistence
      - "/path-to-your-git-repo/:/data" # required to update your repository
volumes:
  restart-compose-tsnet: {}
```

On the Tailscale side, you many need to add an ACL or grant to allow devices inside your tailnet to access the Funnel address.
