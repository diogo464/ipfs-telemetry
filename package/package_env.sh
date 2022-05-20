#!/usr/bin/env bash

if [ -z "$PACKAGE_ARCH" ]; then
    echo "need to set PACKAGE_ARCH" 
    exit 1
fi

export PACKAGE_ARCH
export PACKAGE_VERSION=${PACKAGE_VERSION:-"$(git describe --tags --abbrev=0)"}
export PACKAGE_OUTPUT_DIR=${PACKAGE_OUTPUT_DIR:-"bin/"}
