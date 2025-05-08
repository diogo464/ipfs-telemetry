#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

TAG=ghcr.io/diogo464/ipfs-telemetry/backend:latest

docker build -t $TAG -f backend.Containerfile .
docker push $TAG
