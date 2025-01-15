# Development cluster template

The development cluster template will create an LXC or OCI container running a haproxy server for the cluster load balancer endpoint. The load balancer endpoint will be the IP address of the haproxy container.

> **WARNING**: The load balancer container is a single point of failure for the control plane of the workload cluster, therefore should only be used for development or evaluation purposes.

## Requirements

1. The instance network is reachable by the management controller.

## Configuration

```bash
{{#include ../../../../../templates/development.rc }}
```

## Cluster Template

```yaml
{{#include ../../../../../templates/development.yaml }}
```
