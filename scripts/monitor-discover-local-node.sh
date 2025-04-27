#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

bin/ipfs id | nats publish discovery
