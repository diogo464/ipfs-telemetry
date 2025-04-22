#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

pushd monitor
    make
popd

mkdir -p bin/
cp monitor/monitor bin/monitor
