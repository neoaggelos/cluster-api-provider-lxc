# Kini Quick Start

This document explains how to install `kini` and use it to spin up test Kubernetes clusters on Incus.

## Table Of Contents

<!-- toc -->

## Installation

`kini` binaries are published in [GitHub releases](https://github.com/lxc/cluster-api-provider-incus/releases/latest).

The simplest way to install `kini` is by using the installer script:

```bash
curl https://capn.linuxcontainers.org/get-kini.sh | bash -x
```

> **NOTE**: It is highly recommended that you inspect the script before executing.

Alternatively, you can manually download the binary and install on PATH. For example:

```bash
curl -fsSL https://github.com/lxc/cluster-api-provider-incus/releases/latest/download/kini-linux-amd64
chmod +x kini-linux-amd64
sudo install kini-linux-amd64 /usr/local/bin/kini
```

## Verify installation

Check version of binary (and kind version) using the following command ...

```bash
kini version
```

... which should print something like:

```bash
kini version v0.8.3 (go1.25.5 linux/amd64)
kind version v0.31.0 (default image kindest/node:v1.35.0@sha256:452d707d4862f52530247495d180205e029056831160e22870e37e3f6c1ac31f)
```

## Prepare infrastructure

> **NOTE**: If `docker` is installed on the host, ensure the iptables FORWARD policy is set to ACCEPT. This is required because of how `docker` mangles the iptables rules on the host. Without this, LXC containers will not be able to contact each other.

```bash
sudo iptables -P FORWARD ACCEPT
```

> **NOTE**: You can skip this step if Incus is already installed and configured (i.e. you have a default network and storage pool).

{{#tabs name:"install" tabs:"Incus,Canonical LXD" }}
{{#tab Incus }}

Install `incus` from latest stable version:

```bash
curl https://pkgs.zabbly.com/get/incus-stable | sudo bash -x
```

Initialize Incus with a default bridge network and local disk:

```bash
sudo incus admin init --auto
sudo incus network set incusbr0 ipv6.address=none
```

`kini` will use the same configuration file that the `incus` CLI uses. Validate configuration with:

```bash
sudo kini setup validate-config --v=5
```

{{#/tab }}
{{#tab Canonical LXD }}

Install `lxd`:

```bash
sudo snap install lxd --channel 5.21/stable
```

Initialize LXD with a default bridge network and local disk:

```bash
sudo lxd init --auto
sudo lxc network set lxdbr0 ipv6.address=none
```

`kini` will use the same configuration file that the `lxc` CLI uses. Validate configuration with:

```bash
sudo kini setup validate-config --v=5
```

{{#/tab }}
{{#/tabs }}

## Create a cluster

To create a single node Kubernetes cluster with default configuration ...

```bash
sudo kini kind create cluster
```

... which should produce output like this:

```bash
Creating cluster "kind" ...
 ✓ Ensuring node image (kindest/node:v1.35.0) 🖼
 ✓ Preparing nodes 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing CNI 🔌
 ✓ Installing StorageClass 💾
Set kubectl context to "kind-kind"
You can now use your cluster with:

kubectl cluster-info --context kind-kind

Have a nice day! 👋
```

Then, `kubectl` can be used to access the cluster. Example output:

```bash
# kubectl get node
NAME                 STATUS   ROLES           AGE     VERSION
kind-control-plane   Ready    control-plane   3m41s   v1.35.0
```

To delete the cluster, you can use:

```bash
sudo kini kind delete cluster
```
