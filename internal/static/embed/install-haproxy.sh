#!/bin/sh -xeu

# Usage:
#  $ /opt/cluster-api/install-haproxy.sh

set -xeu

apt update
apt install haproxy -y --no-install-recommends
