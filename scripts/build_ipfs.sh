#!/usr/bin/env bash

if [ "$#" -ne "2" ]; then
    echo "usage: $0 <os> <arch>"
    exit 1
fi

OS=$1
ARCH=$2
EXT=""

if [ "$OS" = "windows" ]; then
    EXT=".exe"
fi

export CGO_ENABLED=0
export GOOS=$1
export GOARCH=$2
export IPFS_BIN="ipfs_$OS-$ARCH$EXT"
make ipfs
