# KubeVIP cluster template

The kube-vip cluster-template will create a static pod running [kube-vip] in the control plane nodes. The control plane endpoint will be the VIP address managed by kube-vip.

## Table Of Contents

<!-- toc -->

## Requirements

1. A free IP address in the workload cluster network.
2. The management cluster can connect to the VIP address (to be able to connect to the workload cluster).

## Configuration

```bash
{{#include ../../../../../templates/cluster-template-kube-vip.rc }}
```

## Generate cluster

```bash
clusterctl generate cluster example-cluster -i incus --flavor kube-vip
```

## Cluster Template

```yaml
{{#include ../../../../../templates/cluster-template-kube-vip.yaml }}
```

<!-- links -->
[kube-vip]: https://kube-vip.io/
