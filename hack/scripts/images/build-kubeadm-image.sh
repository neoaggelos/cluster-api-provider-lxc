#!/bin/bash -xeu

# Examples:
#   $ ./build-kubeadm-image.sh kubeadm-v1.32.1-lxc incus container 24.04 v1.32.1
#   $ ./build-kubeadm-image.sh kubeadm-v1.32.1-lxd-vm lxd virtual-machine 24.04 v1.32.1
#
# Description:
#   Build kubeadm image for target infrastructure, instance type, base image, kubeadm version
#   The image is exported to $1.tar.gz

DIR="$(dirname "$(realpath "$0")")"

OUTPUT="$1"
INFRASTRUCTURE="$2"
MACHINE_TYPE="$3"
UBUNTU_VERSION="$4"
KUBERNETES_VERSION="$5"

BUILDER="${BUILDER:-kubeadm-builder}"

case "$INFRASTRUCTURE" in
  incus)
    CLIENT="incus"
    IMAGE="images:ubuntu/$UBUNTU_VERSION/cloud"
    ;;
  lxd)
    CLIENT="lxc"
    IMAGE="ubuntu:$UBUNTU_VERSION"
    ;;
  *)
    echo "Unknown infrastructure $INFRASTRUCTURE. Must be one of incus, lxd."
    exit 1
    ;;
esac

case "$MACHINE_TYPE" in
  virtual-machine)
    EXTRA_LAUNCH_ARGS="--vm -d root,size=5GiB"
    ;;
  container)
    EXTRA_LAUNCH_ARGS=""
    ;;
  *)
    echo "Unknown machine type $MACHINE_TYPE. Must be one of virtual-machine, container."
    exit 1
    ;;
esac

"$CLIENT" launch "$IMAGE" "$BUILDER" $EXTRA_LAUNCH_ARGS
while ! "$CLIENT" exec "$BUILDER" -- echo hi; do
  sleep 3;
done

cat "${DIR}/stages/00-prerun.sh" | "$CLIENT" exec "$BUILDER" -- bash -s -- "$KUBERNETES_VERSION"
cat "${DIR}/stages/50-install-kubeadm.sh" | "$CLIENT" exec "$BUILDER" -- bash -s -- "$KUBERNETES_VERSION"
cat "${DIR}/stages/90-postrun.sh" | "$CLIENT" exec "$BUILDER" -- bash -s -- "$KUBERNETES_VERSION"
cat "${DIR}/stages/95-postrun-kubeadm.sh" | "$CLIENT" exec "$BUILDER" -- bash -s -- "$KUBERNETES_VERSION"
cat "${DIR}/stages/99-cleanup.sh" | "$CLIENT" exec "$BUILDER" -- bash -s -- "$KUBERNETES_VERSION"

"$CLIENT" stop "$BUILDER"
while ! "$CLIENT" list "$BUILDER" | grep -q STOPPED; do
  sleep 3;
done

case "$INFRASTRUCTURE" in
  incus)  "$CLIENT" snapshot create "$BUILDER" v0   ;;
  lxd)    "$CLIENT" snapshot "$BUILDER" v0          ;;
esac

"$CLIENT" publish "$BUILDER/v0" --alias "$OUTPUT"
"$CLIENT" image export "$OUTPUT" "$OUTPUT"
