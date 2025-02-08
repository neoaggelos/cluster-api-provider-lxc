#!/bin/bash -xeu

# Usage:
#  $ /opt/cluster-api-lxc/95-postrun-kubeadm.sh

set -xeu

# manifest
(
  set -x

  date
  containerd --version
  runc --version
  crictl --version
  kubelet --version
  kubeadm version -o yaml
  kubectl version -o yaml --client
  crictl images
) 2>&1 | tee /opt/manifest.txt
