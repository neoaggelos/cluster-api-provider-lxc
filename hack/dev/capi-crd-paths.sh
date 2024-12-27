#!/bin/bash

# Print paths of ClusterAPI CRDs from GOPATH.

DIR="$(realpath "$(dirname "${0}")")"

cd "${DIR}/../../"
crd_paths="$(go env GOPATH)/pkg/mod/$(cat go.mod | grep 'sigs.k8s.io/cluster-api v' | tr -d '\t' | tr ' ' '@')/config/crd/bases"
[ -d "${crd_paths}" ] && echo "${crd_paths}"
