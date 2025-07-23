# Build kubeadm images

This how-to describes the process of building a custom base image for your infrastructure, instead of having to rely on the [default simplestreams server](../../reference/default-simplestreams-server.md).

The `kubeadm` image will be used to launch cluster nodes.

## Table Of Contents

<!-- toc -->

## Requirements

- A locally configured Incus or Canonical LXD instance. The `image-builder` utility will use the default client credentials.
- Go 1.23.0+

## Build `image-builder` binary

First, clone the cluster-api-provider-incus source repository:

```bash
git clone https://github.com/lxc/cluster-api-provider-incus
```

Then, build the `image-builder` binary with:

```bash
make image-builder
```

## Build `kubeadm` image for containers

Use `./bin/image-builder kubeadm --help` for a list of all available options.

```bash
./bin/image-builder kubeadm --v=4 --output image-kubeadm.tar.gz \
  --kubernetes-version v1.33.0 \
  --base-image ubuntu:24.04 \
  --image-alias kubeadm/v1.33.0/ubuntu/24.04
```

This will build a kubeadm image for Kubernetes v1.33.0, save it with alias `kubeadm/v1.33.0/ubuntu/24.04` and also export it to `image-kubeadm.tar.gz`.

## Build `kubeadm` image for virtual machines

```bash
./bin/image-builder kubeadm --v=4 --output image-kubeadm-kvm.tar.gz \
  --kubernetes-version v1.33.0 \
  --base-image ubuntu:24.04 \
  --image-alias kubeadm/v1.33.0/ubuntu/24.04/kvm \
  --instance-type virtual-machine
```

This will build a kubeadm image for Kubernetes v1.33.0, save it with alias `kubeadm/v1.33.0/ubuntu/24.04/kvm` and also export it to `image-kubeadm-kvm.tar.gz`.

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
+----------------------------------+--------------+--------+---------------------------------------------------+--------------+-----------------+------------+-----------------------+
|              ALIAS               | FINGERPRINT  | PUBLIC |                    DESCRIPTION                    | ARCHITECTURE |      TYPE       |    SIZE    |      UPLOAD DATE      |
+----------------------------------+--------------+--------+---------------------------------------------------+--------------+-----------------+------------+-----------------------+
| kubeadm/v1.33.0/ubuntu/24.04     | 8960df007461 | yes    | kubeadm v1.33.0 ubuntu noble amd64 (202504280150) | x86_64       | CONTAINER       | 742.47MiB  | 2025/04/28 01:50 EEST |
+----------------------------------+--------------+--------+---------------------------------------------------+--------------+-----------------+------------+-----------------------+
| kubeadm/v1.33.0/ubuntu/24.04/kvm | 501df06be7a4 | yes    | kubeadm v1.33.0 ubuntu noble amd64 (202504280156) | x86_64       | VIRTUAL-MACHINE | 1005.12MiB | 2025/04/28 01:57 EEST |
+----------------------------------+--------------+--------+---------------------------------------------------+--------------+-----------------+------------+-----------------------+
```

## Use the image in LXCMachineTemplate

### Using the default cluster templates

When using the example [Cluster Templates](../../reference/templates/), you need to set:

```bash
export CONTROL_PLANE_MACHINE_TYPE=container         # 'container' or 'virtual-machine'
export WORKER_MACHINE_TYPE=container                # must match type of built image

export LXC_IMAGE_NAME=kubeadm/v1.33.0/ubuntu/24.04  # exported image alias name
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
        name: kubeadm/v1.33.0/ubuntu/24.04
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
      instanceType: virtual-machine
      image:
        name: kubeadm/v1.33.0/ubuntu/24.04/kvm
```
