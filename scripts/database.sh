#!/usr/bin/env sh

CONTAINER_NAME="monitor-influxdb"

if [[ "$1" == "up" ]]; then
    if podman container exists $CONTAINER_NAME ; then
        echo database already up
    else
        podman run --rm -itd --name $CONTAINER_NAME \
            -e DOCKER_INFLUXDB_INIT_USERNAME="user" \
            -e DOCKER_INFLUXDB_INIT_PASSWORD="password" \
            -e DOCKER_INFLUXDB_INIT_ORG="adc" \
            -e DOCKER_INFLUXDB_INIT_BUCKET="telemetry" \
            -v $CONTAINER_NAME:/var/lib/influxdb2 \
            -p 8086:8086 \
            docker.io/influxdb:latest
    fi
elif [[ "$1" == "down" ]]; then
    podman stop $CONTAINER_NAME
fi
