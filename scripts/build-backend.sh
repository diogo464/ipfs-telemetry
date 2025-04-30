#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

mkdir -p bin
pushd backend
    go build -o ../bin/backend ./cmd
popd
