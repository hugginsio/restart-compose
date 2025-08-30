# restart-compose

A webhook listener tool for automatically updating Git repositories and restarting Docker Compose stacks.

## Overview

`restart-compose` is a Go application that monitors Docker Compose stacks and provides webhook functionality for automated deployments. It reads configuration from a YAML file and integrates with the Docker API to manage container stacks.

## Features

- ✅ **Configuration Management**: Loads stack definitions from `.restart-compose.yaml`
- ✅ **Stack Validation**: Verifies that compose files exist on the filesystem
- ✅ **Docker Integration**: Connects to Docker daemon and monitors running containers
- ✅ **Stack Information**: Displays detailed information about configured stacks and their services
- 🚧 **Webhook Server**: (In development) Will listen for webhooks to trigger updates

## Configuration

Create a `.restart-compose.yaml` file in your working directory:

```yaml
stacks:
  - "portainer/compose.yaml"
  - "golang/compose.yaml"
```

The paths are relative to the directory containing the configuration file.

## Usage

```bash
# Run the application
go run restart-compose.go

# Or build and run
go build -o restart-compose
./restart-compose
```

## Output Example

```
Starting restart-compose webhook listener...
Found configuration file: /path/to/.restart-compose.yaml
Loaded 2 stack(s) from configuration
Connected to Docker daemon

==================================================
DOCKER COMPOSE STACK INFORMATION
==================================================

=== Stack: portainer ===
Path: /path/to/portainer/compose.yaml
Directory: /path/to/portainer
Exists: true
Services: None running

=== Stack: golang ===
Path: /path/to/golang/compose.yaml
Directory: /path/to/golang
Exists: true
Services (3):
  - golang-app: Up 2 hours (running)
  - redis: Up 2 hours (running)
  - postgres: Up 2 hours (running)

==================================================
Summary: 2 stack(s) configured, 2 exist on filesystem
```

## Requirements

- Go 1.25.0 or later
- Docker daemon running and accessible
- Docker Compose files in the configured locations

## Dependencies

- `github.com/docker/docker` - Docker client API
- `gopkg.in/yaml.v3` - YAML configuration parsing

## Development Status

This project is currently in development. The following features are implemented:

- [x] Configuration file loading
- [x] Stack validation and discovery
- [x] Docker daemon integration
- [x] Container status reporting
- [ ] Webhook server implementation
- [ ] Git repository updating
- [ ] Automatic stack restarting
- [ ] Authentication and security

## License

BSD-3-Clause License. See LICENSE file for details.

## Author

Kyle Huggins