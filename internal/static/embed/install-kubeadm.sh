#!/bin/bash -xeu

# Usage:
#  $ /opt/cluster-api/install-kubeadm.sh v1.32.1

set -xeu

KUBERNETES_VERSION="${KUBERNETES_VERSION:-$1}"            # https://dl.k8s.io/release/stable.txt or https://dl.k8s.io/release/stable-1.32.txt
CNI_PLUGINS_VERSION="${CNI_PLUGINS_VERSION:-v1.6.2}"      # https://github.com/containernetworking/plugins
CRICTL_VERSION="${CRICTL_VERSION:-v1.32.0}"               # https://github.com/kubernetes-sigs/cri-tools
CONTAINERD_VERSION="${CONTAINERD_VERSION:-v1.7.27}"       # https://github.com/containerd/containerd
RUNC_VERSION="${RUNC_VERSION:-v1.2.5}"                    # https://github.com/opencontainers/runc, must match https://raw.githubusercontent.com/containerd/containerd/${CONTAINERD_VERSION}/script/setup/runc-version

KUBELET_SERVICE='
# Sourced from: https://raw.githubusercontent.com/kubernetes/release/v0.16.2/cmd/krel/templates/latest/kubelet/kubelet.service

[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=https://kubernetes.io/docs/
Wants=network-online.target
After=network-online.target

[Service]
ExecStart=/usr/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
'

KUBELET_SERVICE_KUBEADM_DROPIN_CONFIG='
# Sourced from: https://raw.githubusercontent.com/kubernetes/release/v0.16.2/cmd/krel/templates/latest/kubeadm/10-kubeadm.conf

# Note: This dropin only works with kubeadm and kubelet v1.11+
[Service]
Environment="KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf"
Environment="KUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yaml"
# This is a file that "kubeadm init" and "kubeadm join" generates at runtime, populating the KUBELET_KUBEADM_ARGS variable dynamically
EnvironmentFile=-/var/lib/kubelet/kubeadm-flags.env
# This is a file that the user can use for overrides of the kubelet args as a last resort. Preferably, the user should use
# the .NodeRegistration.KubeletExtraArgs object in the configuration files instead. KUBELET_EXTRA_ARGS should be sourced from this file.
EnvironmentFile=-/etc/sysconfig/kubelet
ExecStart=
ExecStart=/usr/bin/kubelet $KUBELET_KUBECONFIG_ARGS $KUBELET_CONFIG_ARGS $KUBELET_KUBEADM_ARGS $KUBELET_EXTRA_ARGS
'

CONTAINERD_CONFIG='
version = 2

[plugins."io.containerd.grpc.v1.cri"]
  stream_server_address = "127.0.0.1"
  stream_server_port = "10010"
  enable_unprivileged_ports = true
  enable_unprivileged_icmp = true
  device_ownership_from_security_context = false
  sandbox_image = "registry.k8s.io/pause:3.10"
  enable_selinux = false

[plugins."io.containerd.grpc.v1.cri".containerd]
  snapshotter = "overlayfs"
  disable_snapshot_annotations = true

[plugins."io.containerd.grpc.v1.cri".cni]
  bin_dir = "/opt/cni/bin"
  conf_dir = "/etc/cni/net.d"

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
  runtime_type = "io.containerd.runc.v2"

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
  SystemdCgroup = true

[plugins."io.containerd.grpc.v1.cri".registry]
  config_path = "/etc/containerd/certs.d"
'

CONTAINERD_UNPRIVILEGED_CONFIG='
version = 2

[plugins."io.containerd.grpc.v1.cri"]
  stream_server_address = "127.0.0.1"
  stream_server_port = "10010"
  enable_unprivileged_ports = true
  enable_unprivileged_icmp = true
  device_ownership_from_security_context = false
  sandbox_image = "registry.k8s.io/pause:3.10"
  enable_selinux = false

  ## unprivileged
  disable_apparmor = true
  disable_hugetlb_controller = true
  restrict_oom_score_adj = true

[plugins."io.containerd.grpc.v1.cri".containerd]
  snapshotter = "overlayfs"
  disable_snapshot_annotations = true

[plugins."io.containerd.grpc.v1.cri".cni]
  bin_dir = "/opt/cni/bin"
  conf_dir = "/etc/cni/net.d"

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
  runtime_type = "io.containerd.runc.v2"

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc.options]
  ## unprivileged
  SystemdCgroup = false

[plugins."io.containerd.grpc.v1.cri".registry]
  config_path = "/etc/containerd/certs.d"
'

CONTAINERD_SERVICE='
[Unit]
Description=containerd container runtime
Documentation=https://containerd.io
After=network.target local-fs.target

[Service]
#uncomment to enable the experimental sbservice (sandboxed) version of containerd/cri integration
#Environment="ENABLE_CRI_SANDBOXES=sandboxed"
ExecStartPre=-/sbin/modprobe overlay
ExecStart=/usr/bin/containerd

Type=notify
Delegate=yes
KillMode=process
Restart=always
RestartSec=5
# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNPROC=infinity
LimitCORE=infinity
LimitNOFILE=infinity
# Comment TasksMax if your systemd version does not supports it.
# Only systemd 226 and above support this version.
TasksMax=infinity
OOMScoreAdjust=-999

[Install]
WantedBy=multi-user.target
'

CONTAINERD_CONFIGURE_UNPRIVILEGED_MODE='#!/bin/bash -xeu

set -xeu

ln -sf config.unprivileged.toml /etc/containerd/config.toml
systemctl restart containerd
'

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
if ! systemd-detect-virt --container --quiet 2>/dev/null; then
  modprobe br_netfilter
  echo br_netfilter | tee /etc/modules-load.d/br_netfilter.conf
fi

# apt install requirements
apt update
apt install curl iptables ethtool --no-install-recommends --yes
if [ "$KUBERNETES_VERSION" "<" "v1.32" ]; then
  apt install conntrack --no-install-recommends --yes
fi

# runc
curl -L "https://github.com/opencontainers/runc/releases/download/${RUNC_VERSION}/runc.${ARCH}" -o /usr/bin/runc
chmod +x /usr/bin/runc
cp /usr/bin/runc /usr/sbin/runc

# containerd
mkdir -p /etc/containerd
curl -L "https://github.com/containerd/containerd/releases/download/${CONTAINERD_VERSION}/containerd-static-${CONTAINERD_VERSION#v}-linux-${ARCH}.tar.gz" | tar -C /usr -xz
if [ ! -f /etc/containerd/config.toml ]; then
  echo "${CONTAINERD_CONFIG}" | tee /etc/containerd/config.default.toml
  echo "${CONTAINERD_UNPRIVILEGED_CONFIG}" | tee /etc/containerd/config.unprivileged.toml
  ln -sf config.default.toml /etc/containerd/config.toml
fi
if ! systemctl list-unit-files containerd.service &>/dev/null; then
  echo "${CONTAINERD_SERVICE}" | tee /usr/lib/systemd/system/containerd.service
fi
systemctl enable containerd.service
systemctl start containerd.service

# containerd unprivileged mode
echo "${CONTAINERD_CONFIGURE_UNPRIVILEGED_MODE}" | tee /opt/containerd-configure-unprivileged-mode.sh
chmod +x /opt/containerd-configure-unprivileged-mode.sh

# cni plugins
mkdir -p /opt/cni/bin
curl -L "https://github.com/containernetworking/plugins/releases/download/${CNI_PLUGINS_VERSION}/cni-plugins-linux-${ARCH}-${CNI_PLUGINS_VERSION}.tgz" | tar -C /opt/cni/bin -xz

# crictl
curl -L "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${ARCH}.tar.gz" | tar -C /usr/bin -xz
echo 'runtime-endpoint: unix:///run/containerd/containerd.sock' | tee -a /etc/crictl.yaml

# kubernetes binaries
curl -L --remote-name-all "https://dl.k8s.io/release/${KUBERNETES_VERSION}/bin/linux/${ARCH}/kubeadm" -o /usr/bin/kubeadm
curl -L --remote-name-all "https://dl.k8s.io/release/${KUBERNETES_VERSION}/bin/linux/${ARCH}/kubelet" -o /usr/bin/kubelet
curl -L --remote-name-all "https://dl.k8s.io/release/${KUBERNETES_VERSION}/bin/linux/${ARCH}/kubectl" -o /usr/bin/kubectl
chmod +x /usr/bin/kubeadm /usr/bin/kubelet /usr/bin/kubectl

# kubelet service
mkdir -p /usr/lib/systemd/system/kubelet.service.d
if ! systemctl list-unit-files kubelet.service &>/dev/null; then
  echo "${KUBELET_SERVICE}" | tee /usr/lib/systemd/system/kubelet.service
  echo "${KUBELET_SERVICE_KUBEADM_DROPIN_CONFIG}" | tee /usr/lib/systemd/system/kubelet.service.d/10-kubeadm.conf
fi
systemctl enable kubelet.service

# pull images
kubeadm config images pull --kubernetes-version "${KUBERNETES_VERSION}"
