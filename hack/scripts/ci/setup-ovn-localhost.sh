#!/bin/bash -xeu

# install and configure single-node OVN
sudo apt install ovn-host ovn-central -y
sudo ovs-vsctl set open_vswitch . external_ids:ovn-encap-ip="127.0.0.1"
sudo ovs-vsctl set open_vswitch . external_ids:ovn-encap-type="geneve"
sudo ovs-vsctl set open_vswitch . external_ids:ovn-remote="unix:/run/ovn/ovnsb_db.sock"
