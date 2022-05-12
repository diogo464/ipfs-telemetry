#!/usr/bin/env bash

cd /project || exit 1
pacman --noconfirm -Sy git base-devel traceroute
useradd -m builder || exit 1
mkdir /output && chown -R builder /output

runuser -u builder -- git config --global --add safe.directory /project
PACKAGE_OUTPUT_DIR=/output runuser -u builder -- bash -c 'package/package_archlinux.sh'

cp /output/* bin/
