#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

pushd crawler
    make
popd

mkdir -p bin/
cp crawler/crawler bin/crawler
