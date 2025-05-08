#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

TAG=ghcr.io/diogo464/ipfs-telemetry/ipfs:latest

docker build -t $TAG -f ipfs.Containerfile .
docker push $TAG
