#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

bin/ipfs id | bin/nats publish discovery
