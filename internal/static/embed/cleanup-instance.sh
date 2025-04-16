#!/bin/sh -xeu

# Usage:
#  $ /opt/cluster-api-lxc/99-cleanup.sh

set -xeu

apt-get autoremove && apt-get clean && apt-get autoclean
rm -rf \
  /var/lib/apt/lists \
  /var/cache/apt \
  /home/ubuntu/.cache \
  /home/ubuntu/.config \
  /home/ubuntu/.gnupg \
  /home/ubuntu/.ssh \
  /home/ubuntu/.sudo_as_admin_successful \
  /home/ubuntu/.bash_history \
  /root/.cache \
  /root/.config \
  /root/.gnupg \
  /root/.ssh \
  /root/.sudo_as_admin_successful \
  /root/.bash_history \
  /var/lib/swcatalog \
  /usr/share/doc \
  /usr/share/man

if which cloud-init; then
  cloud-init clean --machine-id --seed --logs
fi
