#!/bin/bash -xeu

DIR="$(dirname "$(realpath "$0")")"

if ! which incus; then
  curl https://pkgs.zabbly.com/get/incus-stable | sudo bash -x
fi

# get node IP address
ip_address="$(ip -o route get to 1.1.1.1 | sed -n 's/.*src \([0-9.]\+\).*/\1/p')"

# Initialize Incus
sudo incus admin init --auto --network-address "$ip_address"

# Generate client certificate and key, trust certificate
if ! incus remote switch local-https; then
  incus remote generate-certificate
  sudo incus config trust add-certificate ~/.config/incus/client.crt
  incus remote add local-https "https://$(sudo incus config get core.https_address)" --accept-certificate
  incus remote switch local-https
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
  server: '$(echo -n "https://$(incus config get core.https_address)" | base64 -w0)'
  server-crt: '$(cat ~/.config/incus/servercerts/local-https.crt | base64 -w0)'
  client-crt: '$(cat ~/.config/incus/client.crt | base64 -w0)'
  client-key: '$(cat ~/.config/incus/client.key | base64 -w0)'
" | tee "${DIR}/../../../lxc-secret.yaml"

# Setup local Incus daemon for e2e tests
CLI=incus "${DIR}/setup-e2e.sh"
