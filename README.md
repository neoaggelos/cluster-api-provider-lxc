# Kubernetes Cluster API Provider Incus

[![End to End Tests](https://github.com/lxc/cluster-api-provider-incus/actions/workflows/test-e2e-full.yml/badge.svg)](https://github.com/lxc/cluster-api-provider-incus/actions/workflows/test-e2e-full.yml)
[![End to End Tests [Conformance]](https://github.com/lxc/cluster-api-provider-incus/actions/workflows/test-e2e-conformance.yml/badge.svg)](https://github.com/lxc/cluster-api-provider-incus/actions/workflows/test-e2e-conformance.yml)
[![Unit Tests](https://github.com/lxc/cluster-api-provider-incus/actions/workflows/test.yml/badge.svg)](https://github.com/lxc/cluster-api-provider-incus/actions/workflows/test.yml)
[![Deploy GitHub Pages](https://github.com/lxc/cluster-api-provider-incus/actions/workflows/docs.yml/badge.svg)](https://github.com/lxc/cluster-api-provider-incus/actions/workflows/docs.yml)

Kubernetes-native declarative infrastructure for [Incus](https://linuxcontainers.org/incus/introduction/), [Canonical LXD](https://canonical.com/lxd) and [Canonical MicroCloud](https://canonical.com/microcloud).

## What is the Cluster API Provider Incus

[Cluster API](https://cluster-api.sigs.k8s.io) is a Kubernetes sub-project focused on providing declarative APIs and tooling to simplify provisioning, upgrading, and operating multiple Kubernetes clusters.

`cluster-api-provider-incus` (CAPN) is an Infrastructure Provider for Cluster API, which enables deploying clusters on infrastructure operated by [Incus](https://linuxcontainers.org/incus/introduction/), [Canonical LXD](https://canonical.com/lxd) and [Canonical MicroCloud](https://canonical.com/microcloud).

The provider can be used in single-node development environments for evaluation and testing, but also work with multi-node clusters to deploy and manage production Kubernetes clusters.

## Documentation

Please refer to our [book](https://lxc.github.io/cluster-api-provider-incus) for in-depth documentation.

## Quick Start

See [Quick Start](https://lxc.github.io/cluster-api-provider-incus/tutorial/quick-start.html) to launch a cluster on a single-node development environment.

## Features

- Supports [Incus](https://linuxcontainers.org/incus/introduction/), [Canonical LXD](https://canonical.com/lxd) and [Canonical MicroCloud](https://canonical.com/microcloud).
- Support for [kube-vip](https://lxc.github.io/cluster-api-provider-incus/reference/templates/kube-vip.html) (production), [OVN network load balancers](https://lxc.github.io/cluster-api-provider-incus/reference/templates/ovn.html) or simple [haproxy containers](https://lxc.github.io/cluster-api-provider-incus/reference/templates/development.html) (development) for the cluster load balancer.
- [Default simplestreams server](https://lxc.github.io/cluster-api-provider-incus/reference/default-simplestreams-server.html) with pre-built kubeadm images.
- Supports virtual machines or LXC containers for the cluster machines. Automatically manages the [profile](https://lxc.github.io/cluster-api-provider-incus/reference/profile/kubeadm.html) for Kubernetes to work in LXC containers.
- Can be used for local development similar to CAPD for quickly iterating on custom bootstrap and control-plane providers, e.g. K3s, Canonical Kubernetes, etc.

## Project Roadmap

### v0.8.0

Rough steps for version v0.8.0:

- [ ] Add `kini` command line tool, re-using building blocks from [kind](https://kind.sigs.k8s.io).
- [ ] Use `kini` for quick start guide and e2e tests.
- [x] Private initial alpha testing.
- [x] Cloud provider node patch to link Machines with workload cluster Nodes.
- [x] Test with both [Incus](https://linuxcontainers.org/incus/introduction/) and [Canonical LXD](https://canonical.com/lxd).
- [x] Start cluster-api-provider-incus book with quick start guide, cluster templates, API reference.
- [x] Publish v0.1.0 release to get initial user feedback.
- [x] Add e2e tests using the cluster-api testing framework.
- [x] Add PR blocking CI pipelines.
- [x] Publish v0.2.0 release with v1alpha2 APIs.
- [x] Add e2e tests for cluster upgrades.
- [x] Explore clusters with ClusterTopology=true (clusterclass), also allows us to run all existing ClusterAPI e2e tests like Autoscaler, etc.
- [x] Write developer guide.
- [x] Support unprivileged containers.
- [x] Support configurable machine placement for production clusters.
- [x] Extend e2e suite with tests for all cluster-template types (kvm, unprivileged containers, kube-vip, ovn)
- [x] Add self-hosted e2e test.
- [x] Implement `kind` instance types (using OCI containers with the kindest/node images from the kind project).
- [x] Gather initial user feedback.
- [ ] Improve generated API reference documentation.
- [ ] Add cluster-templates for 3rd party providers, e.g. [Canonical Kubernetes](https://github.com/canonical/cluster-api-k8s).
- [ ] Write documentation with common troubleshooting steps.
- [ ] Write documentation with common cluster deployment scenarios.

### $Future

- [ ] Add to default list of providers supported by ClusterAPI.
- [ ] Improve API validations and possibly API conformance tests.
- [x] Add CI to build kubeadm images for the default simplestreams server. Pushing will remain manual for now.
- [x] Decide on project OWNERSHIP and testing infrastructure (part of LXC org).
- [x] Refactor `internal/incus` package and improve consistency and log levels across the code.
- [ ] Split cloud provider node patch to external cloud-provider-incus project.

## Getting involved and contributing

The `cluster-api-provider-incus` project would love your suggestions, contributions and help! The maintainers can be contacted at any time to learn mode about how to get involved.

Remember that there are numerous effective ways to contribute to the project: raise a pull request to fix a bug, improve test coverage, improve existing documentation or even participate in GitHub issues. We want your help!

Please refer to the [developer guide](https://lxc.github.io/cluster-api-provider-incus/howto/developer-guide.html) in order to get started with setting up a local environment for development and testing.
