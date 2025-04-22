#!/usr/bin/env -S bash -x
set -e

TMPDIR=$(mktemp -d)
pushd $TMPDIR
    curl -L https://github.com/nats-io/natscli/releases/download/v0.2.2/nats-0.2.2-linux-amd64.zip -o nats.zip
    unzip nats.zip
popd
mv $(find $TMPDIR -name nats) bin/nats
rm -rf "$TMPDIR"
