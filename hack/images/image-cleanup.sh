#!/bin/bash -xeu

# Usage:
#  $ /opt/cluster-api-lxc/99-cleanup.sh

set -xeu

rm -rfv /var/lib/apt/lists
apt-get autoremove && apt-get clean && apt-get autoclean
rm -rfv /home/ubuntu/.cache /home/ubuntu/.config /home/ubuntu/.gnupg /home/ubuntu/.ssh /home/ubuntu/.sudo_as_admin_successful /home/ubuntu/.bash_history
rm -rfv /root/.cache /root/.config /root/.gnupg /root/.ssh /root/.sudo_as_admin_successful /root/.bash_history

if which cloud-init; then
  cloud-init clean --machine-id --seed --logs
fi
