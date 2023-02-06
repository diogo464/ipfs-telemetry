#!/bin/sh

cd $(dirname $0)
helm upgrade -f values/ipfs-telemetry.yaml ipfs-telemetry ../chart/