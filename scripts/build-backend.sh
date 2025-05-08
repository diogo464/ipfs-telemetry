#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

mkdir -p bin
pushd backend
    CGO_ENABLED=0 go build -o ../bin/backend ./cmd
popd
