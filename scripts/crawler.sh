#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

PROMETHEUS_ADDRESS="0.0.0.0:9091" \
    bin/backend crawler
