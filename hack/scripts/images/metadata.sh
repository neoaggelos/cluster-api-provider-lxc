## build images

export KUBERNETES_VERSION=v1.32.1

## directories

mkdir -p build                # build artifacts
mkdir -p staging              # staging directory, set image metadata
mkdir -p out                  # image output directory
mkdir -p simplestreams        # simplestreams index and images

BUILDER=lxc ./scripts/images/build-kubeadm-image.sh build/lxc incus container 24.04 "$KUBERNETES_VERSION"
BUILDER=ivm ./scripts/images/build-kubeadm-image.sh build/ivm incus virtual-machine 24.04 "$KUBERNETES_VERSION"
BUILDER=lvm ./scripts/images/build-kubeadm-image.sh build/lvm lxd   virtual-machine 24.04 "$KUBERNETES_VERSION"

BUILDER=haproxy OUT_IMAGE=build/haproxy ./scripts/images/haproxy.sh

## haproxy

mkdir -p staging/haproxy
(
  set -x
  cd staging/haproxy
  sudo tar xzfp ../../build/haproxy.tar.gz

  sed -i 's,expiry_date: .*,expiry_date: 3741161000,' ./metadata.yaml
  sed -i 's,os: .*,os: haproxy,' ./metadata.yaml
  sed -i 's,variant: .*,variant: ubuntu,' ./metadata.yaml
  sed -i 's,name: .*,name: haproxy-ubuntu-24.04,' ./metadata.yaml
  sed -i 's,description: .*,description: haproxy-ubuntu-24.04,' ./metadata.yaml

  sudo tar czf ../../out/haproxy.tar.gz metadata.yaml templates rootfs
  sudo chown $(id -u):$(id -g) ../../out/haproxy.tar.gz
)

## lxc

mkdir -p staging/lxc
(
  set -x
  cd staging/lxc
  sudo tar xzfp ../../build/lxc.tar.gz

  sed -i '/expiry_date:/d' ./metadata.yaml
  sed -i 's,os: .*,os: kubeadm,' ./metadata.yaml
  sed -i "s,release: .*,release: ${KUBERNETES_VERSION}," ./metadata.yaml
  sed -i 's,variant: .*,variant: ubuntu,' ./metadata.yaml
  sed -i "s,name: .*,name: kubeadm-${KUBERNETES_VERSION}-ubuntu-24.04," ./metadata.yaml
  sed -i "s,description: .*,description: kubeadm-${KUBERNETES_VERSION}-ubuntu-24.04," ./metadata.yaml

  sudo tar czf ../../out/lxc.tar.gz metadata.yaml templates rootfs
  sudo chown $(id -u):$(id -g) ../../out/lxc.tar.gz
)

## lvm
mkdir -p staging/lvm
(
  set -x
  cd staging/lvm
  sudo tar xzfp ../../build/lvm.tar.gz

  sudo sed -i '/expiry_date:/d' ./metadata.yaml
  sudo sed -i 's,os: .*,os: kubeadm,' ./metadata.yaml
  sudo sed -i "s,release: .*,release: ${KUBERNETES_VERSION}," ./metadata.yaml
  sudo sed -i 's,variant: .*,variant: ubuntu,' ./metadata.yaml
  sudo sed -i "s,name: .*,name: kubeadm-${KUBERNETES_VERSION}-ubuntu-24.04," ./metadata.yaml
  sudo sed -i "s,description: .*,description: kubeadm-${KUBERNETES_VERSION}-ubuntu-24.04," ./metadata.yaml
  sudo sed -i 's/properties:/properties:\n    variant: ubuntu/' ./metadata.yaml        # add missing 'variant' image property
  sudo sed -i 's,  architecture: "x86_64",  architecture: amd64,' ./metadata.yaml      # ensure 'architecture' is amd64 (consistent with incus)

  sudo tar czf ../../out/lvm.tar.gz metadata.yaml templates
  sudo cp rootfs.img ../../out/lvm.qcow2

  sudo chown $(id -u):$(id -g) ../../out/lvm.*
  sudo chmod 0644 ../../out/lvm.*
)

## ivm
mkdir -p staging/ivm
(
  set -x
  cd staging/ivm
  sudo tar xzfp ../../build/ivm.tar.gz

  sed -i '/expiry_date:/d' ./metadata.yaml
  sed -i 's,os: .*,os: kubeadm,' ./metadata.yaml
  sed -i "s,release: .*,release: ${KUBERNETES_VERSION}," ./metadata.yaml
  sed -i 's,variant: .*,variant: ubuntu,' ./metadata.yaml
  sed -i "s,name: .*,name: kubeadm-${KUBERNETES_VERSION}-ubuntu-24.04," ./metadata.yaml
  sed -i "s,description: .*,description: kubeadm-${KUBERNETES_VERSION}-ubuntu-24.04," ./metadata.yaml

  sudo tar czf ../../out/ivm.tar.gz metadata.yaml templates
  sudo cp rootfs.img ../../out/ivm.qcow2

  sudo chown $(id -u):$(id -g) ../../out/ivm.*
)

sudo apt install incus-extra -y

### Simplestreams (order is important as we hack sed replace incus tags for LXD compatibility)
(
  set -x
  cd simplestreams

  # LXD container (lxd_combined.tar.gz)
  # Metadata hacks:
  #  ftype 'incus_combined.tar.gz' -> 'lxd_combined.tar.gz'
  incus-simplestreams add ../out/lxc.tar.gz --no-default-alias --alias kubeadm/$KUBERNETES_VERSION --alias kubeadm/$KUBERNETES_VERSION/ubuntu
  incus-simplestreams add ../out/haproxy.tar.gz --no-default-alias --alias haproxy --alias haproxy/ubuntu
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
  incus-simplestreams add ../out/haproxy.tar.gz --no-default-alias --alias haproxy --alias haproxy/ubuntu

  # Incus virtual machine (incus.tar.xz + disk-kvm.img)
  incus-simplestreams add ../out/ivm.tar.gz ../out/ivm.qcow2 --no-default-alias --alias kubeadm/$KUBERNETES_VERSION --alias kubeadm/$KUBERNETES_VERSION/ubuntu
)
