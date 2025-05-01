#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

mkdir -p data/pg
podman run -d --name pg --network host \
    -e POSTGRES_HOST_AUTH_METHOD=trust \
    -e POSTGRES_PASSWORD=telemetry \
    docker.io/postgres:latest
