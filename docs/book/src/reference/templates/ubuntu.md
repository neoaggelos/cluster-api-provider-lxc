# Ubuntu cluster template

The ubuntu cluster template is the same as the [development](./development.md) cluster template, but works with an upstream Ubuntu 24.04 instance and installs kubeadm during cloud-init.

> **WARNING**: The load balancer container is a single point of failure for the control plane of the workload cluster, therefore should only be used for development or evaluation purposes.

> **WARNING**: cloud-init will download all binaries on all nodes while deploying the cluster. This is wasteful and will take longer than using a base image.

## Table Of Contents

<!-- toc -->

## Requirements

1. The instance network is reachable by the management controller.
2. Instances can reach GitHub to pull binaries and install kubeadm.

## Configuration

```bash
{{#include ../../../../../templates/ubuntu.rc }}
```

## Generate cluster

```bash
clusterctl generate cluster example-cluster \
    --from https://neoaggelos.github.io/cluster-api-provider-lxc/static/v0.1/release/templates/ubuntu.yaml
```

## Cluster Template

```yaml
{{#include ../../../../../templates/ubuntu.yaml }}
```
