#!/bin/bash -e

# Usage: $0 <output-dir> <version>

DIR="$(dirname "$(realpath "$0")")"
OUT="$(realpath "${1}")"

# build for each platform
GOOS=linux  GOARCH=amd64 "${DIR}/build.sh" "${OUT}/kini-linux-amd64" "${2}"
GOOS=linux  GOARCH=arm64 "${DIR}/build.sh" "${OUT}/kini-linux-arm64" "${2}"
GOOS=darwin GOARCH=arm64 "${DIR}/build.sh" "${OUT}/kini-darwin-arm64" "${2}"
