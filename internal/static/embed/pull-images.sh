#!/bin/sh -xeu

# Usage:
#  $ /opt/cluster-api/pull-images.sh [IMAGE] [IMAGE] ....

set -xeu

for image in "${@}"; do
  crictl pull "${image}"
done
