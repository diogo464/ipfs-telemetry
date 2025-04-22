#!/usr/bin/env -S bash -x

DATA_DIR="./data/grafana"
PROVISIONING_DIR="./grafana/provisioning"
DASHBOARD_DIR="./grafana/dashboards"

mkdir -p "$DATA_DIR"

# uses port 3000

podman run --name grafana -d \
    -v "$DATA_DIR:/var/lib/grafana:z,U" \
    -v "$PROVISIONING_DIR:/etc/grafana/provisioning:z,ro" \
    -v "$DASHBOARD_DIR:/var/lib/grafana/dashboards:z,ro" \
    -e GF_SECURITY_ADMIN_PASSWORD="telemetry" \
    -e GF_AUTH_ANONYMOUS_ENABLED="true" \
    --network host \
    docker.io/grafana/grafana:latest

