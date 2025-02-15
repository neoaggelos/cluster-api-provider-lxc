#!/bin/bash -xeu

KUBERNETES_VERSION=v1.32.1
### Simplestreams (order is important as we hack sed replace incus tags for LXD compatibility)
(
  cd simplestreams

  # LXD container (lxd_combined.tar.gz)
  # Metadata hacks:
  #  ftype 'incus_combined.tar.gz' -> 'lxd_combined.tar.gz'
  incus-simplestreams add ../out/lxc.tar.gz --no-default-alias --alias kubeadm/$KUBERNETES_VERSION --alias kubeadm/$KUBERNETES_VERSION/ubuntu
  #incus-simplestreams add ../haproxy.tar.gz --no-default-alias --alias haproxy --alias haproxy/ubuntu
  sed -i 's,"incus_combined.tar.gz","lxd_combined.tar.gz",g' ./streams/v1/images.json

  # LXD virtual machine (lxd.tar.gz + disk1.img)
  # Metadata hacks:
  # hack up metadata for lxd
  #  ftype 'disk-kvm.img' -> 'disk1.img'
  #  ftype 'incus.tar.gz' -> 'lxd.tar.xz'
  #  fhash 'combined_disk-kvm-img_sha256' -> 'combined_disk1-img_sha256'
  incus-simplestreams add ../out/lvm.tar.gz ../out/lvm.qcow2 --no-default-alias --alias kubeadm/$KUBERNETES_VERSION --alias kubeadm/$KUBERNETES_VERSION/ubuntu
  sed -i 's,"disk-kvm.img","disk1.img",g' streams/v1/images.json
  sed -i 's,"incus.tar.xz","lxd.tar.xz",g' streams/v1/images.json
  sed -i 's,"combined_disk-kvm-img_sha256","combined_disk1-img_sha256",g' streams/v1/images.json

  # Incus container (incus_combined.tar.gz)
  incus-simplestreams add ../out/lxc.tar.gz --no-default-alias --alias kubeadm/$KUBERNETES_VERSION --alias kubeadm/$KUBERNETES_VERSION/ubuntu
  #incus-simplestreams add ../haproxy.tar.gz --no-default-alias --alias haproxy --alias haproxy/ubuntu

  # Incus virtual machine (incus.tar.xz + disk-kvm.img)
  incus-simplestreams add ../out/ivm.tar.gz ../out/ivm.qcow2 --no-default-alias --alias kubeadm/$KUBERNETES_VERSION --alias kubeadm/$KUBERNETES_VERSION/ubuntu
)
