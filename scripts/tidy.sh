#!/usr/bin/env sh

ORIGIN=$(pwd)
for d in ./ third_party/go-ipfs/ third_party/go-bitswap third_party/go-libp2p-kad-dht third_party/go-libp2p-kbucket ; do
    cd $d
    go mod tidy
    cd $ORIGIN
done
