#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

TAG=ghcr.io/diogo464/ipfs-telemetry/grafana:latest

docker build -t $TAG -f grafana.Containerfile .
docker push $TAG
