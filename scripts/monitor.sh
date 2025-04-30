#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

PROMETHEUS_ADDRESS="0.0.0.0:9090" \
MONITOR_COLLECT_INTERVAL="5s" \
    bin/backend monitor
