#!/usr/bin/env -S bash
set -e

for id in $(podman container list --noheading | grep kubo | cut -d' ' -f1); do
  podman exec $id ipfs id | nats publish monitor.discover
done
