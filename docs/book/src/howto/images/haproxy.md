# Build haproxy images

This how-to describes the process of building a custom base image for your infrastructure, instead of having to rely on the [default simplestreams server](../../reference/default-simplestreams-server.md).

The `haproxy` image will be used for the cluster load balancer when using the development cluster template.

We will go over the steps of launching a builder instance with the appropriate base image, installing haproxy, cleaning up and publishing a snapshot of the image, as well as steps for using it.

## Table Of Contents

<!-- toc -->

## Launch builder instance

We must start a `container` instance with the base Operating System that we want to have. Ubuntu 24.04 and Debian 12 have been tested and are known to work.

> **NOTE**: The images published on [default simplestreams server](../../reference/default-simplestreams-server.md) are based on Ubuntu 24.04.

{{#tabs name:"launch" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

Launch a new Virtual Machine using the Ubuntu 24.04 image from [https://images.linuxcontainers.org/](https://images.linuxcontainers.org/).

```bash
incus launch images:ubuntu/24.04 haproxy-builder
```

Or you can use Debian 12 as base image:

```bash
incus launch images:debian/12 haproxy-builder
```

{{#/tab }}

{{#tab Canonical LXD }}

Launch a new Virtual Machine using the Ubuntu 24.04 image from [https://cloud-images.ubuntu.com/releases/](https://cloud-images.ubuntu.com/releases/).

```bash
lxc launch ubuntu:24.04 haproxy-builder
```

{{#/tab }}

{{#/tabs }}

## Pre-run commands

Launch any commands you might need _before_ installing haproxy.

## Install haproxy

```bash
{{#include ../../static/v0.1/install-haproxy.sh }}
```

{{#tabs name:"install" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
curl https://lxc.github.io/cluster-api-provider-incus/static/v0.1/install-haproxy.sh | incus exec haproxy-builder -- bash
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
curl https://lxc.github.io/cluster-api-provider-incus/static/v0.1/install-haproxy.sh | lxc exec haproxy-builder -- bash
```

{{#/tab }}

{{#/tabs }}

## Post-run commands

Launch any commands you might need _after_ installing haproxy and adjusting the host configuration.

## Clean-up

We use the script below to cleanup package archives, deb packages, bash history files and local user configurations. Most importantly, we also reset `cloud-init` status, so that we can use it as a base image.

```bash
{{#include ../../static/v0.1/image-cleanup.sh }}
```

{{#tabs name:"cleanup" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
curl https://lxc.github.io/cluster-api-provider-incus/static/v0.1/image-cleanup.sh | incus exec haproxy-builder -- bash
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
curl https://lxc.github.io/cluster-api-provider-incus/static/v0.1/image-cleanup.sh | lxc exec haproxy-builder -- bash
```

{{#/tab }}

{{#/tabs }}

## Stop instance and publish snapshot

At this point, our image root filesystem is ready. Final steps are to shutdown the instance and publish:

{{#tabs name:"publish" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
incus stop haproxy-builder
incus snapshot create haproxy-builder snapshot-0

# publish snapshot as image, using alias "haproxy/ubuntu/24.04"
incus publish haproxy-builder/snapshot-0 --alias haproxy/ubuntu/24.04
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
lxc stop haproxy-builder
lxc snapshot haproxy-builder snapshot-0

# publish snapshot as image, using alias "haproxy/ubuntu/24.04"
lxc publish haproxy-builder/snapshot-0 --alias haproxy/ubuntu/24.04
```

{{#/tab }}

{{#/tabs }}

## Check image

{{#tabs name:"list" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
incus image list haproxy
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
lxc image list haproxy
```

{{#/tab }}

{{#/tabs }}

The output should look similar to this:

```bash
+---------+--------------+--------+-------------------------------------+--------------+-----------+-----------+----------------------+
|  ALIAS  | FINGERPRINT  | PUBLIC |             DESCRIPTION             | ARCHITECTURE |   TYPE    |   SIZE    |     UPLOAD DATE      |
+---------+--------------+--------+-------------------------------------+--------------+-----------+-----------+----------------------+
| haproxy | 464c5e12f184 | no     | Ubuntu noble amd64 (20250115_07:42) | x86_64       | CONTAINER | 133.90MiB | 2025/01/15 23:19 EET |
+---------+--------------+--------+-------------------------------------+--------------+-----------+-----------+----------------------+
```

## Use the image in LXCCluster

Set `spec.loadBalancer.instanceSpec.image.name` on the LXCCluster resource of your workload cluster. When launching the cluster, this will now use our custom image to provision the load balancer.

```yaml,hidelines=#
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: LXCCluster
metadata:
  name: example-cluster
spec:
  secretRef:
    name: example-secret
  loadBalancer:
    lxc:
      instanceSpec:
        #flavor: c1-m1
        #profiles: [default]
        image:
          name: haproxy
```
