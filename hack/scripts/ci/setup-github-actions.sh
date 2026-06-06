#!/bin/bash -xeu

# NOTE(neoaggelos/2025-12-16): Cleanup GitHub Actions runner unused tools
#
# runner@runnervm6qbrg:~/work/test-ci/test-ci$ date
# Mon Dec 15 23:49:43 UTC 2025
# runner@runnervm6qbrg:~/work/test-ci/test-ci$ df -h /
# Filesystem      Size  Used Avail Use% Mounted on
# /dev/root        72G   58G   15G  80% /

if [ "${GITHUB_ACTIONS:=}" == "true" ]; then
  # Print available space before cleanup
  df -h /

  # These finish fast and give back around 2GB (used: 58GB -> 56GB)
  time sudo apt purge llvm*
  time sudo apt autoremove

  # NOTE(neoaggelos/2025-12-16): removing these takes about 1-2 minutes
  # and gives back around 39GB (used: 56GB -> 17GB)
  time echo \
    /home/runner/.dotnet \
    /home/runner/.rustup \
    /opt/az \
    /opt/google \
    /opt/hostedtoolcache/CodeQL \
    /opt/hostedtoolcache/PyPy \
    /opt/hostedtoolcache/Python \
    /opt/microsoft \
    /opt/pipx \
    /usr/lib/firefox \
    /usr/lib/google-cloud-sdk \
    /usr/lib/jvm \
    /usr/lib/llvm* \
    /usr/local/.ghcup \
    /usr/local/aws* \
    /usr/local/bin/azcopy \
    /usr/local/bin/bicep \
    /usr/local/bin/cmake* \
    /usr/local/bin/helm \
    /usr/local/bin/minikube \
    /usr/local/bin/pulumi* \
    /usr/local/bin/packer \
    /usr/local/bin/stack \
    /usr/local/julia* \
    /usr/local/lib/android/sdk/build-tools \
    /usr/local/lib/android/sdk/cmake \
    /usr/local/lib/android/sdk/cmdline-tools \
    /usr/local/lib/android/sdk/extras \
    /usr/local/lib/android/sdk/ndk \
    /usr/local/share/chromium \
    /usr/local/share/powershell \
    /usr/share/az* \
    /usr/share/dotnet \
    /usr/share/gradle* \
    /usr/share/kotlinc \
    /usr/share/man \
    /usr/share/miniconda \
    /usr/share/swift | xargs -n1 -t time sudo rm -rf

  # Print available space after cleanup
  df -h /

  # NOTE(neoaggelos/2026-06-06): /etc/containers/registries.conf uses v1 config format, which breaks
  # skopeo 1.23.0. Example error:
  # failed to wait for PullImage operation: Failed getting remote image info: Failed to run: skopeo --insecure-policy inspect docker://docker.io/kindest/node:v1.35.0 --no-tags: exit status 1 (time=\"2026-06-06T12:27:40Z\" level=fatal msg=\"Error parsing image name \\\"docker://docker.io/kindest/node:v1.35.0\\\": loading registries configuration: loading registries configuration \\\"/etc/containers/registries.conf\\\": registries.conf must be in v2 format but is in v1\")
  cat /etc/containers/registries.conf
  sudo rm /etc/containers/registries.conf
fi
