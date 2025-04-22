#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

pushd kubo
    make build
popd

mkdir -p bin/
cp kubo/cmd/ipfs/ipfs bin/ipfs
