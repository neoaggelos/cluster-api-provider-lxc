# Injected Files

CAPN will always inject the following files on launched instances (through the use of [optional instance templates](https://linuxcontainers.org/incus/docs/main/reference/image_format/#templates-optional)):

## Table Of Contents

<!-- toc -->

## List of files

| File path | Nodes | Usage |
| -|-|-|
| [`/opt/cluster-api/install-kubeadm.sh`](#install-kubeadmsh) | all | Can be used to install kubeadm on the instance, e.g. if using stock Ubuntu images. |
| [`/opt/cluster-api/kube-flannel.yaml`](#kube-flannelyaml) | control-plane | Default manifest that can be used to deploy kube-flannel on the cluster. |
| [`/opt/cluster-api/kube-vip.yaml`](#kube-vipyaml) | control-plane | Default static pod manifest that can be used to deploy kube-vip on the cluster. |
| [`/opt/cluster-api/kube-vip.hosts`](#kube-viphosts) | control-plane | Hosts file for the default kube-vip configuration. |
| [`/opt/cluster-api/kube-proxy-config-lxc.yaml`](#kube-proxy-config-lxcyaml) | control-plane | KubeProxyConfiguration for LXC instances. Appended to the InitConfiguration of kubeadm. |

### install-kubeadm.sh

```bash
# Path: /opt/cluster-api/install-kubeadm.sh

{{#include ../../../../internal/static/embed/install-kubeadm.sh }}
```

### kube-flannel.yaml

#### Instance Configuration

- `user.capn.kube-flannel.pod-network-cidr`: Set by CAPN based on the networkConfiguration of the Cluster object.

```yaml
# Path: /opt/cluster-api/kube-flannel.yaml

{{#include ../../../../internal/static/embed/kube-flannel.tpl }}
```

### kube-vip.yaml

#### Instance Configuration

- `user.capn.kube-vip.host`: Set by CAPN based on the control plane endpoint of the Cluster object.
- `user.capn.kube-vip.port`: Set by CAPN based on the control plane endpoint of the Cluster object.
- `user.capn.kube-vip.interface`: Can be set by the user to specify the interface name that kube-vip should use.
- `user.capn.kube-vip.image`: Can be set by the user to specify the kube-vip image to use.

```yaml
# Path: /opt/cluster-api/kube-vip.yaml

{{#include ../../../../internal/static/embed/kube-vip.tpl }}
```

### kube-vip.hosts

```ini
# Path: /opt/cluster-api/kube-vip.hosts

{{#include ../../../../internal/static/embed/kube-vip-hosts.tpl }}
```

### kube-proxy-config-lxc.yaml

```yaml
# Path: /opt/cluster-api/kube-proxy-config-lxc.yaml

{{#include ../../../../internal/static/embed/kube-proxy-config.tpl }}
```
