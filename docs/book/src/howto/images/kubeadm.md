# Build kubeadm images

This how-to describes the process of building a custom base image for your infrastructure, instead of having to rely on the [default simplestreams server](../../reference/default-simplestreams-server.md).

The `kubeadm` image will be used to launch cluster nodes.

We will go over the steps of launching a builder instance with the appropriate base image, installing Kubernetes and other necessary tools, cleaning up and publishing a snapshot of the image, as well as steps for using it.

## Table Of Contents

<!-- toc -->

## Launch builder instance

We must start a builder instance that matches the instance type (`container` or `virtual-machine`) and base Operating System that we want to have. Ubuntu 24.04 and Debian 12 have been tested and are known to work.

> **NOTE**: The images published on [default simplestreams server](../../reference/default-simplestreams-server.md) are based on Ubuntu 24.04.

The one requirement for our base image is that it comes with `cloud-init`, since ClusterAPI needs `cloud-init` to provision instances.

{{#tabs name:"launch" tabs:"Incus (VM),Incus (Container),Canonical LXD (VM), Canonical LXD (Container)" }}

{{#tab Incus (VM) }}

Launch a new Virtual Machine using the Ubuntu 24.04 image from [https://images.linuxcontainers.org/](https://images.linuxcontainers.org/). Make sure to use the `cloud` variant, which supports cloud-init.

```bash
incus launch images:ubuntu/24.04 kubeadm-builder
```

Or you can use Debian 12 as base image:

```bash
incus launch images:debian/12 kubeadm-builder
```

{{#/tab }}

{{#tab Incus (Container) }}

Launch a new Container using the Ubuntu 24.04 image from [https://images.linuxcontainers.org/](https://images.linuxcontainers.org/). Make sure to use the `cloud` variant, which supports cloud-init.

```bash
incus launch images:ubuntu/24.04 kubeadm-builder
```

Or you can use Debian 12 as base image:

```bash
incus launch images:debian/12 kubeadm-builder
```
{{#/tab }}


{{#tab Canonical LXD (VM) }}

Launch a new Virtual Machine using the Ubuntu 24.04 image from [https://cloud-images.ubuntu.com/releases/](https://cloud-images.ubuntu.com/releases/). The image comes with support for cloud-init.

```bash
lxc launch ubuntu:24.04 kubeadm-builder --vm
```

{{#/tab }}

{{#tab Canonical LXD (Container) }}

Launch a new Container using the Ubuntu 24.04 image from [https://cloud-images.ubuntu.com/releases/](https://cloud-images.ubuntu.com/releases/). The image comes with support for cloud-init.

```bash
lxc launch ubuntu:24.04 kubeadm-builder
```

{{#/tab }}

{{#/tabs }}

The steps to build the kubeadm image are the same for both Container and Virtual Machine images.

## Pre-run commands

Launch any commands you might need _before_ installing the Kubernetes binaries and set any host configuration. This might include installing extra packages and binaries depending on your requirements, configuring default registry mirrors based on your target environment, etc.

## Install Kubernetes

We will use the following script to install runc, containerd, cni-plugins, crictl and Kubernetes binaries on the instance. Feel free to adjust component versions as required.

```bash
{{#include ../../static/v0.1/install-kubeadm.sh }}
```

Run the script on the instance using the command below. Make sure to specify the Kubernetes version you want to use, e.g. `v1.31.4`:

{{#tabs name:"install" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
curl https://lxc.github.io/cluster-api-provider-incus/static/v0.1/install-kubeadm.sh | incus exec kubeadm-builder -- bash -s -- v1.31.4
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
curl https://lxc.github.io/cluster-api-provider-incus/static/v0.1/install-kubeadm.sh | lxc exec kubeadm-builder -- bash -s -- v1.31.4
```

{{#/tab }}

{{#/tabs }}

## Post-run commands

Launch any commands you might need _after_ installing the Kubernetes and adjusting the host configuration. This might include patching component configurations, or pulling extra OCI images.

## Clean-up

We use the script below to cleanup package archives, deb packages, bash history files and local user configurations. Most importantly, we also reset `cloud-init` status, so that we can use it as a base image.


```bash
{{#include ../../static/v0.1/image-cleanup.sh }}
```

{{#tabs name:"cleanup" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
curl https://lxc.github.io/cluster-api-provider-incus/static/v0.1/image-cleanup.sh | incus exec kubeadm-builder -- bash
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
curl https://lxc.github.io/cluster-api-provider-incus/static/v0.1/image-cleanup.sh | lxc exec kubeadm-builder -- bash
```

{{#/tab }}

{{#/tabs }}

## Stop instance and publish snapshot

At this point, our image root filesystem is ready. Final steps are to shutdown the instance and publish:

{{#tabs name:"publish" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
incus stop kubeadm-builder
incus snapshot create kubeadm-builder snapshot-0

# publish snapshot as image, using alias "kubeadm/v1.31.4/ubuntu/24.04"
incus publish kubeadm-builder/snapshot-0 --alias kubeadm/v1.31.4/ubuntu/24.04
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
lxc stop kubeadm-builder
lxc snapshot kubeadm-builder snapshot-0

# publish snapshot as image, using alias "kubeadm/v1.31.4/ubuntu/24.04"
lxc publish kubeadm-builder/snapshot-0 --alias kubeadm/v1.31.4/ubuntu/24.04
```

{{#/tab }}

{{#/tabs }}

## Check image

{{#tabs name:"list" tabs:"Incus,Canonical LXD" }}

{{#tab Incus }}

```bash
incus image list kubeadm
```

{{#/tab }}

{{#tab Canonical LXD }}

```bash
lxc image list kubeadm
```

{{#/tab }}

{{#/tabs }}

The output should look similar to this:

```bash
+------------------------------+--------------+--------+-----------------------------------+--------------+-----------+-----------+----------------------+
|            ALIAS             | FINGERPRINT  | PUBLIC |            DESCRIPTION            | ARCHITECTURE |   TYPE    |   SIZE    |     UPLOAD DATE      |
+------------------------------+--------------+--------+-----------------------------------+--------------+-----------+-----------+----------------------+
| kubeadm/v1.31.4/ubuntu/24.04 | 5053ee6cac52 | no     | ubuntu noble amd64 (202501120531) | x86_64       | CONTAINER | 668.13MiB | 2025/01/12 07:24 EET |
+------------------------------+--------------+--------+-----------------------------------+--------------+-----------+-----------+----------------------+
```

## Use the image in LXCMachineTemplate

### Using the default cluster templates

When using the example [Cluster Templates](../../reference/templates/), you need to set:

```bash
export CONTROL_PLANE_MACHINE_TYPE=container         # 'container' or 'virtual-machine'
export WORKER_MACHINE_TYPE=container                # must match type of built image

export LXC_IMAGE_NAME=kubeadm/v1.31.4/ubuntu/24.40  # exported image alias name
```

### Editing LXCImageTemplate manually

The image name must be set on the `spec.image.name` field on the LXCMachineTemplate resources of your workload cluster. When launching the cluster, this will now use our custom image to provision the instances.

Make sure to set `.spec.instanceType` to `container` or `virtual-machine` accordingly (depending on the kind of image you built), for example:

```yaml,hidelines=#
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: LXCMachineTemplate
metadata:
  name: "example-control-plane"
spec:
  template:
    spec:
      #flavor: c2-m4
      #profiles: [default]
      instanceType: container
      image:
        name: kubeadm/v1.31.4/ubuntu/24.04
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha2
kind: LXCMachineTemplate
metadata:
  name: "example-md-0"
spec:
  template:
    spec:
      #flavor: c2-m4
      #profiles: [default]
      instanceType: container
      image:
        name: kubeadm/v1.31.4/ubuntu/24.04
```
