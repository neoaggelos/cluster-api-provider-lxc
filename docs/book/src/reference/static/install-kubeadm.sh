#!/bin/bash -xeu

# Usage:
#  $ /opt/cluster-api-lxc/50-install-kubeadm.sh v1.32.0

set -xeu

KUBERNETES_VERSION="${KUBERNETES_VERSION:-$1}"
CNI_PLUGINS_VERSION="${CNI_PLUGINS_VERSION:-v1.6.0}"
CRICTL_VERSION="${CRICTL_VERSION:-v1.31.0}"
KREL_VERSION="${KREL_VERSION:-v0.16.2}"
CONTAINERD_VERSION="${CONTAINERD_VERSION:-v1.7.24}"
RUNC_VERSION="${RUNC_VERSION:-v1.2.3}"    # must match https://raw.githubusercontent.com/containerd/containerd/${CONTAINERD_VERSION}/script/setup/runc-version

CONTAINERD_CONFIG_URL="${CONTAINERD_CONFIG_URL:-"https://neoaggelos.github.io/cluster-api-provider-lxc/reference/static/config.toml"}"
CONTAINERD_SERVICE_URL="${CONTAINERD_SERVICE_URL:-"https://neoaggelos.github.io/cluster-api-provider-lxc/reference/static/containerd.service"}"

# infer ARCH
ARCH="$(uname -m)"
if uname -m | grep -q x86_64; then ARCH=amd64; fi
if uname -m | grep -q aarch64; then ARCH=arm64; fi

# sysctl
echo net.ipv4.ip_forward=1 | tee /etc/sysctl.d/99-clusterapi.conf
echo fs.inotify.max_user_instances=8192 | tee -a /etc/sysctl.d/99-clusterapi.conf
echo fs.inotify.max_user_watches=524288 | tee -a /etc/sysctl.d/99-clusterapi.conf
sysctl --system

# kernel
if ! systemd-detect-virt -c -q 2>/dev/null; then
  modprobe br_netfilter
  echo br_netfilter | tee /etc/modules-load.d/br_netfilter.conf
fi

# apt install requirements
apt update
apt install curl iptables ethtool --no-install-recommends -y

# runc
curl -L "https://github.com/opencontainers/runc/releases/download/${RUNC_VERSION}/runc.${ARCH}" -o /usr/bin/runc
chmod +x /usr/bin/runc
cp /usr/bin/runc /usr/sbin/runc

# containerd
mkdir -p /etc/containerd
curl -L "https://github.com/containerd/containerd/releases/download/${CONTAINERD_VERSION}/containerd-static-${CONTAINERD_VERSION#v}-linux-${ARCH}.tar.gz" | tar -C /usr -xz
if [ ! -f /etc/containerd/config.toml ]; then
  curl -L "${CONTAINERD_CONFIG_URL}" | tee /etc/containerd/config.toml
fi
if ! systemctl list-unit-files containerd.service &>/dev/null; then
  curl -L "${CONTAINERD_SERVICE_URL}" | tee /usr/lib/systemd/system/containerd.service
fi
systemctl enable containerd.service
systemctl start containerd.service

# cni plugins
mkdir -p /opt/cni/bin
curl -L "https://github.com/containernetworking/plugins/releases/download/${CNI_PLUGINS_VERSION}/cni-plugins-linux-${ARCH}-${CNI_PLUGINS_VERSION}.tgz" | tar -C /opt/cni/bin -xz

# crictl
curl -L "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${ARCH}.tar.gz" | tar -C /usr/bin -xz

# kubernetes binaries
curl -L --remote-name-all "https://dl.k8s.io/release/${KUBERNETES_VERSION}/bin/linux/${ARCH}/kubeadm" -o /usr/bin/kubeadm
curl -L --remote-name-all "https://dl.k8s.io/release/${KUBERNETES_VERSION}/bin/linux/${ARCH}/kubelet" -o /usr/bin/kubelet
curl -L --remote-name-all "https://dl.k8s.io/release/${KUBERNETES_VERSION}/bin/linux/${ARCH}/kubectl" -o /usr/bin/kubectl
chmod +x /usr/bin/kubeadm /usr/bin/kubelet /usr/bin/kubectl

# kubelet service
mkdir -p /usr/lib/systemd/system/kubelet.service.d
if ! systemctl list-unit-files kubelet.service &>/dev/null; then
  curl -sSL "https://raw.githubusercontent.com/kubernetes/release/${KREL_VERSION}/cmd/krel/templates/latest/kubelet/kubelet.service" | tee /usr/lib/systemd/system/kubelet.service
  curl -sSL "https://raw.githubusercontent.com/kubernetes/release/${KREL_VERSION}/cmd/krel/templates/latest/kubeadm/10-kubeadm.conf" | tee /usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf
fi
systemctl enable kubelet.service

# pull images
kubeadm config images pull
