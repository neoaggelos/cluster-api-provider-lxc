# Kubernetes Cluster API Provider LXC

Kubernetes-native declarative infrastructure for [Incus](https://linuxcontainers.org/incus/introduction/), [Canonical LXD](https://canonical.com/lxd) and [Canonical MicroCloud](https://canonical.com/microcloud).

## What is the Cluster API Provider LXC

[Cluster API](https://cluster-api.sigs.k8s.io) is a Kubernetes sub-project focused on providing declarative APIs and tooling to simplify provisioning, upgrading, and operating multiple Kubernetes clusters.

`cluster-api-provider-lxc` (CAPL) is an Infrastructure Provider for Cluster API, which enables deploying clusters on infrastructure operated by [Incus](https://linuxcontainers.org/incus/introduction/), [Canonical LXD](https://canonical.com/lxd) and [Canonical MicroCloud](https://canonical.com/microcloud).

The provider can be used in single-node development environments for evaluation and testing, but also work with multi-node clusters to deploy and manage production Kubernetes clusters.

## Documentation

Please refer to our [book](https://neoaggelos.github.io/cluster-api-provider-lxc) for in-depth documentation.

## Quick Start

See [Quick Start](./tutorial/quick-start.md) to launch a cluster on a single-node development environment.

## Features

- Supports [Incus](https://linuxcontainers.org/incus/introduction/), [Canonical LXD](https://canonical.com/lxd) and [Canonical MicroCloud](https://canonical.com/microcloud).
- Support for [kube-vip](./reference/templates/kube-vip.md) (production), [OVN network load balancers](./reference/templates/ovn.md) or simple [haproxy containers](./reference/templates/development.md) (development) for the cluster load balancer.
- [Default simplestreams server](./reference/default-simplestreams-server.md) with pre-built kubeadm images.
- Supports virtual machines or LXC containers for the cluster machines. Automatically manages the [profile](./reference/profile/kubeadm.md) for Kubernetes to work in LXC containers.
- Can be used for local development similar to CAPD for quickly iterating on custom bootstrap and control-plane providers, e.g. K3s, Canonical Kubernetes, etc.

## Project Roadmap

### v0.2.0

Rough steps for version v0.2.0:

- [x] Private initial alpha testing.
- [x] Cloud provider node patch to link Machines with workload cluster Nodes.
- [x] Test with both [Incus](https://linuxcontainers.org/incus/introduction/) and [Canonical LXD](https://canonical.com/lxd).
- [x] Start cluster-api-provider-lxc book with quick start guide, cluster templates, API reference.
- [x] Publish v0.1.0 release to get initial user feedback.
- [ ] Write documentation with setting up a development environment.
- [ ] Write documentation with common troubleshooting steps.
- [ ] Write documentation with common cluster deployment scenarios.
- [ ] Add e2e tests using the cluster-api testing framework.
- [ ] Add PR blocking CI pipelines.
- [ ] Add cluster-templates for 3rd party providers, e.g. [Canonical Kubernetes](https://github.com/canonical/cluster-api-k8s).
- [ ] Gather initial user feedback.

### $Future

- [ ] Improve API validations and possibly API conformance tests.
- [ ] Add CI to build and push kubeadm and haproxy images to the default simplestreams server.
- [ ] Decide on project OWNERSHIP and testing infrastructure.
- [ ] Split cloud provider node patch to external cloud-provider-incus project.
- [ ] Explore clusters with ClusterTopology=true (clusterclass).
- [ ] Refactor `internal/incus` package and improve consistency and log levels across the code.
- [ ] Add to default list of providers supported by ClusterAPI.

## Getting involved and contributing

The `cluster-api-provider-lxc` project would love your suggestions, contributions and help! The maintainers can be contacted at any time to learn mode about how to get involved.

Remember that there are numerous effective ways to contribute to the project: raise a pull request to fix a bug, improve test coverage, improve existing documentation or even participate in GitHub issues. We want your help!
