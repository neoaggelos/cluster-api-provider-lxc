# Compatibility Notes

## Table Of Contents

<!-- toc -->

## Overview

Internally, `kind` invokes `docker` CLI commands to perform necessary actions. `kini` works by presenting a shim `docker` executable, which instead interacts with Incus/LXD, rather than the Docker daemon.

Most `docker` CLI actions can be 1-1 mapped to Incus actions, but the platforms present a few differences, which the shim layer needs to handle. These are described in this document.

If interested, the implementation for `docker` shim can found in [cmd/exp/kini/docker](https://github.com/lxc/cluster-api-provider-incus/tree/main/cmd/exp/kini/docker).

## Command support matrix

| Command                   | Supported                                                                                                                                                                                                                                                 |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind build node-image`   | **NO**  <br><br> `kind build node-image` builds a kindest/node image. It installs Kubernetes binaries and generates the image using `docker commit`. Support to export OCI images in Incus is currently lacking, therefore this feature is not supported. |
| `kind create cluster`     | **YES**                                                                                                                                                                                                                                                   |
| `kind delete cluster`     | **YES**                                                                                                                                                                                                                                                   |
| `kind export kubeconfig`  | **YES**                                                                                                                                                                                                                                                   |
| `kind export logs`        | **YES** <br><br> Docker related information contains Incus server information instead                                                                                                                                                                     |
| `kind get clusters`       | **YES**                                                                                                                                                                                                                                                   |
| `kind get nodes`          | **YES**                                                                                                                                                                                                                                                   |
| `kind get kubeconfig`     | **YES**                                                                                                                                                                                                                                                   |
| `kind load image-archive` | **YES**                                                                                                                                                                                                                                                   |
| `kind load docker-image`  | **YES(*)**  <br><br> Does not interact with the Docker daemon. Instead, images must be loaded into kini using `kini docker load -i image.tar`                                                                                                             |

## Privileged vs unprivileged containers

By default, `kind` uses privileged containers. In order to use `unprivileged` containers, it detects whether the docker daemon is configured in rootless mode, and adjusts accordingly.

For `kini`, this is instead handled by the environment variable `KINI_UNPRIVILEGED`.
- `KINI_UNPRIVILEGED=false` will spin up privileged containers.
- `KINI_UNPRIVILEGED=true` will spin up unprivileged containers. This is more secure, but known limitations apply. See [Unprivileged containers](../../explanation/unprivileged-containers.md) for more details.

By default, `kini` will launch privileged containers.

## LXC and OCI instances

It is part of `kini`'s mission to support both Incus and Canonical LXD. However, Canonical LXD does not support OCI instances, therefore it is not possible to launch regular kind containers. Instead, `kini` will override the launch command, and spin up LXC instances using the [default kubeadm images of the cluster-api-provider-incus project](../../reference/default-simplestreams-server.md#provided-images).

This is controlled using the `KINI_MODE` environment variable, which accepts the following values:
- `KINI_MODE=oci` always spin up kind instances as OCI containers
- `KINI_MODE=lxc` always spin up kind instances as LXC containers

If not set, `kini` will use the `oci` mode if the platform supports OCI containers (Incus v6.11 or newer). Otherwise, it will fall back to `lxc` mode.

The full support table is:

| Server          | `OCI`   | `LXC`   |
| --------------- | ------- | ------- |
| `Incus`         | **YES** | **YES** |
| `Canonical LXD` | **NO**  | **YES** |

Because of the difference between LXC and OCI containers, please be aware of the following differences:

| Description   | `KINI_MODE=oci`                                                          | `KINI_MODE=lxc`                                                                                       |
| ------------- | ------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------- |
| Haproxy LB    | Uses the `docker.io/kindest/haproxy` image. Overrides the default config | Only supports single-node kind clusters                                                               |
| Kind images   | The well-known `docker.io/kindest/node` images are used                  | Uses the [pre-built kubeadm images](../../reference/default-simplestreams-server.md#provided-images). |
| CNI           | The default `kindnetd` CNI is used                                       | Seeds /kind/manifests/default-cni.yaml with the `kube-flannel` CNI                                    |
| Local Storage | The default `local-path-provisioner` is used                             | Seeds /kind/manifests/default-storage.yaml with `local-path-provisioner`                              |
| Containerd    | `KIND_EXPERIMENTAL_CONTAINERD_SNAPSHOTTER` is always set to `overlayfs`  | Uses the containerd configuration of the pre-built kubeadm images                                     |
| Static files  | `/kind/*` files come from the kindest/node image                         | Seeds `/kind/product_name`, `/kind/product_uuid` and `/kind/version.txt` files manually               |
| Networking    | Does not create a `kind` network, uses the configured one                | Does not create a `kind` network, uses the one already configured                                     |
