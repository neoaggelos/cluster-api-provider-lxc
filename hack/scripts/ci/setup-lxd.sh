#!/bin/bash -xeu

DIR="$(dirname "$(realpath "$0")")"

if ! snap list lxd; then
  sudo apt update
  sudo apt install snapd -y
  sudo snap wait core seed.loaded

  timeout 5m sh -c '
    while ! sudo snap install lxd --channel 5.21/stable; do
      echo retry failed install
    done
  '
fi

sudo snap refresh lxd --channel 5.21/stable

# get node IP address
ip_address="$(ip -o route get to 1.1.1.1 | sed -n 's/.*src \([0-9.]\+\).*/\1/p')"

# Initialize LXD
sudo lxd init --auto --network-address "$ip_address"

# Generate client certificate and key, trust certificate
if ! lxc remote switch local-https; then
  token="$(sudo lxc config trust add --name client | tail -1)"
  lxc remote add local-https "https://$(sudo lxc config get core.https_address)" --token "$token"
  lxc remote set-default local-https
fi

# Do not drop instance traffic
sudo iptables -P FORWARD ACCEPT

# Write Kubernetes secret to top-level dir
echo "
apiVersion: v1
kind: Secret
metadata:
  name: lxc-secret
data:
  project: '$(echo -n default | base64 -w0)'
  server: '$(echo -n "https://$(lxc config get core.https_address)" | base64 -w0)'
  server-crt: '$(cat ~/snap/lxd/common/config/servercerts/local-https.crt | base64 -w0)'
  client-crt: '$(cat ~/snap/lxd/common/config/client.crt | base64 -w0)'
  client-key: '$(cat ~/snap/lxd/common/config/client.key | base64 -w0)'
" | tee "${DIR}/../../../lxc-secret.yaml"

# Setup local LXD daemon for e2e tests
CLI=lxc "${DIR}/setup-e2e-resources.sh"
