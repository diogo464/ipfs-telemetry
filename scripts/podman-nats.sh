#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

# uses port 4222 and 8222

mkdir -p data/nats
podman run -d --name nats --network host \
    -v ./data/nats:/data:z \
    -v ./scripts/nats.conf:/etc/nats.conf:ro \
    docker.io/nats:latest \
    -js -sd /data \
    -c /etc/nats.conf

