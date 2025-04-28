# Build haproxy images

This how-to describes the process of building a custom base image for your infrastructure, instead of having to rely on the [default simplestreams server](../../reference/default-simplestreams-server.md).

The `haproxy` image will be used for the cluster load balancer when using the development cluster template.

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

## Build `haproxy` image

Use `./bin/image-builder haproxy --help` for a list of all available options.

```bash
./bin/image-builder haproxy --v=4 --output image-haproxy.tar.gz \
  --image-alias haproxy/u24 \
  --ubuntu-version 24.04
```

This will build a haproxy image based on Ubuntu 24.04, save it on the server as `haproxy/u24` and also export it to the local file `image-haproxy.tar.gz`

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
+-------------+--------------+--------+------------------------------------+--------------+-----------+-----------+-----------------------+
|    ALIAS    | FINGERPRINT  | PUBLIC |            DESCRIPTION             | ARCHITECTURE |   TYPE    |   SIZE    |      UPLOAD DATE      |
+-------------+--------------+--------+------------------------------------+--------------+-----------+-----------+-----------------------+
| haproxy/u24 | 80aef76c0754 | yes    | haproxy noble amd64 (202504280141) | x86_64       | CONTAINER | 148.15MiB | 2025/04/28 01:41 EEST |
+-------------+--------------+--------+------------------------------------+--------------+-----------+-----------+-----------------------+
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
          name: haproxy/u24
```
