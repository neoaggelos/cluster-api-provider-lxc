# Kubernetes Cluster API Provider LXC

Kubernetes-native declarative infrastructure for [Incus](https://linuxcontainers.org/incus/introduction/), [Canonical LXD](https://canonical.com/lxd) and [Canonical MicroCloud](https://canonical.com/microcloud).

## What is the Cluster API Provider LXC

[Cluster API](https://cluster-api.sigs.k8s.io) is a Kubernetes sub-project focused on providing declarative APIs and tooling to simplify provisioning, upgrading, and operating multiple Kubernetes clusters.

`cluster-api-provider-lxc` is an Infrastructure Provider for Cluster API, which enables deploying clusters on infrastructure operated by [Incus](https://linuxcontainers.org/incus/introduction/), [Canonical LXD](https://canonical.com/lxd) and [Canonical MicroCloud](https://canonical.com/microcloud).

The provider can be used in single-node development environments for evaluation and testing, but also work with multi-node clusters to deploy and manage production Kubernetes clusters.

## Documentation

Please refer to our [book](https://neoaggelos.github.io/cluster-api-provider-lxc) for in-depth documentation.

## Quick Start

See the [Quick Start](./tutorial/quick-start.md) guide to launch a cluster on a single-node development environment.

## Features

- Supports [Incus](https://linuxcontainers.org/incus/introduction/), [Canonical LXD](https://canonical.com/lxd) and [Canonical MicroCloud](https://canonical.com/microcloud).
- Support for [OVN network load balancers](./reference/templates/ovn.md) and [kube-vip](./reference/templates/kube-vip.md) (production), or simple [haproxy containers](./reference/templates/) (development) for the cluster load balancer.
- [Default simplestreams server](./reference/default-simplestreams-server.md) with pre-built kubeadm images.
- Supports virtual machines or LXC containers for the cluster machines. Automatically manages the [profile](./reference/profile/kubeadm.md) for Kubernetes to work in LXC containers.
- Can be used for local development similar to CAPD for quickly iterating on custom bootstrap and control-plane providers, e.g. K3s, Canonical Kubernetes, etc.
