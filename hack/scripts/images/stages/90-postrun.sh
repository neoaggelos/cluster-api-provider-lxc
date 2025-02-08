#!/bin/bash -xeu

# Usage:
#  $ /opt/cluster-api-lxc/90-postrun.sh

set -xeu

declare -a IMAGES=(
  "docker.io/flannel/flannel-cni-plugin:v1.6.0-flannel1"
  "docker.io/flannel/flannel:v0.26.3"
)

for image in "${IMAGES[@]}"; do
  crictl pull "${image}"
done
