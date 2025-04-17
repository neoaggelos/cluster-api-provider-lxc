# Injected Files

CAPN will always inject the following files on launched instances (through the use of [optional instance templates](https://linuxcontainers.org/incus/docs/main/reference/image_format/#templates-optional)):

## Table Of Contents

<!-- toc -->

## List of files

| File path | Nodes | Usage |
| -|-|-|
| [`/opt/cluster-api/install-kubeadm.sh`](#install-kubeadmsh) | all | Can be used to install kubeadm on the instance, e.g. if using stock Ubuntu images. |

### install-kubeadm.sh

```bash
# Path: /opt/cluster-api/install-kubeadm.sh

{{#include ../../../../internal/static/embed/install-kubeadm.sh }}
```
