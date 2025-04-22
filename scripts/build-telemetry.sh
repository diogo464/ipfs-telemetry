#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

pushd telemetry
    make
popd

mkdir -p bin/
cp telemetry/bin/telemetry bin/telemetry
