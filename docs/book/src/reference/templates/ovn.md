# OVN network load balancer cluster template

This cluster template will provision an [OVN network load balancer] to forward traffic to control plane machines on the cluster. The control plane endpoint will be the listen IP address of the network load balancer.

## Table Of Contents

<!-- toc -->

## Requirements

1. Incus configured with OVN.
2. A free IP address in the OVN uplink network.
3. The management cluster can reach the OVN uplink network (to be able to connect to the workload cluster).

## Configuration

> *NOTE*: make sure that the instance profiles will use the OVN network for the instance networking.

```bash
{{#include ../../../../../templates/ovn.rc }}
```

## Generate cluster

```bash
clusterctl generate cluster example-cluster \
    --from https://neoaggelos.github.io/cluster-api-provider-lxc/static/v0.1/release/templates/ovn.yaml
```

## Cluster Template

```yaml
{{#include ../../../../../templates/ovn.yaml }}
```

<!-- links -->
[OVN network load balancer]: https://linuxcontainers.org/incus/docs/main/howto/network_load_balancers/
