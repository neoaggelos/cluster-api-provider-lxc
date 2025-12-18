#!/bin/bash -e

# Usage: $0 <output> <version>

DIR="$(dirname "$(realpath "$0")")"

(
  cd "${DIR}/../../../"
  export CGO_ENABLED=0

  (set -x; go build -o "${1}" \
    -ldflags "-s -w -X=github.com/lxc/cluster-api-provider-incus/cmd/exp/kini/kini.version=$2" \
    ./cmd/exp/kini)

  (cd "$(dirname "${1}")"; sha256sum "$(basename "${1}")" > "${1}.sha256sum")
)
