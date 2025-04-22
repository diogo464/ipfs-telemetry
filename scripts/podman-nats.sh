#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

# uses port 4222 and 8222

mkdir -p data/nats
podman run -d --name nats --network host \
    -v ./data/nats:/data:z \
    docker.io/nats:latest \
    -js -sd /data

