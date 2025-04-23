#!/usr/bin/env -S bash -x

if [ -n "$OUTPUT" ]; then
    OUTPUT=$(realpath $OUTPUT)
else
    OUTPUT="bin/ipfs"
fi

set -e
cd $(dirname $0)/..

pushd kubo
    make build
popd

mkdir -p bin/
cp kubo/cmd/ipfs/ipfs $OUTPUT
