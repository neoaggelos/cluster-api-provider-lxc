#!/bin/bash -eux

###########################################################

DIR="$(dirname "$(realpath "$0")")"

###########################################################

# image alias to create
BUILDER="${BUILDER:-haproxy-builder}"
OUT_IMAGE="${OUT_IMAGE:-haproxy}"

### 1. Incus
CLI=incus
BASE_IMAGE=images:ubuntu/24.04      # one of ["images:ubuntu/24.04/cloud", "images:debian/12/cloud"]

###########################################################

# launch builder instance
$CLI launch "$BASE_IMAGE" "$BUILDER"
while ! $CLI exec "$BUILDER" -- echo hi; do
  sleep 3;
done

###########################################################
# install haproxy and configure

cat "$DIR/stages/50-install-haproxy.sh" | $CLI exec "$BUILDER" -- bash
cat "$DIR/stages/99-cleanup.sh" | $CLI exec "$BUILDER" -- bash

###########################################################
# stop builder and publish as image

$CLI stop "$BUILDER"
$CLI snapshot create "$BUILDER" v0
$CLI publish "$BUILDER/v0" --alias "$OUT_IMAGE"
$CLI image export $OUT_IMAGE $OUT_IMAGE
