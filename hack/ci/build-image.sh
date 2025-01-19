#!/bin/bash -xeu

# Examples:
#   $ ./build-image.sh image.tar.gz incus container 24.04 v1.32.0
#   $ ./build-image.sh image.tar.gz lxd virtual-machine 24.04 v1.32.0

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
    EXTRA_LAUNCH_ARGS="--vm"
    ;;
  container)
    EXTRA_LAUNCH_ARGS=""
    ;;
  *)
    echo "Unknown machine type $MACHINE_TYPE. Must be one of virtual-machine, container."
    exit 1
    ;;
esac

"$CLIENT" launch "$IMAGE" "$BUILDER" $EXTRA_LAUNCH_ARGS -c raw.qemu='-cpu host,-svm'
while ! "$CLIENT" exec "$BUILDER" -- echo hi; do
  "$CLIENT" list
  "$CLIENT" info --show-log "$BUILDER" || true
  "$CLIENT" start "$BUILDER" || true
  sleep 3;
done

cat "${DIR}/../images/install-kubeadm.sh" | "$CLIENT" exec "$BUILDER" -- bash -s -- "$KUBERNETES_VERSION"
cat "${DIR}/../images/image-cleanup.sh" | "$CLIENT" exec "$BUILDER" -- bash

"$CLIENT" stop "$BUILDER"
"$CLIENT" snapshot create "$BUILDER" v0 || "$CLIENT" snapshot "$BUILDER" v0    # "incus snapshot create instance v0" vs "lxc snapshot instance v0"
"$CLIENT" publish "$BUILDER/v0" --alias "$OUTPUT"
"$CLIENT" image export "$OUTPUT" "$OUTPUT"
