#!/bin/bash
set -eo pipefail

GOCC="${GOCC:-go}"

IPFS_PATH="../third_party/go-ipfs/"
IPFS_MODFILE="$IPFS_PATH/go.mod"

TMP="$(mktemp -d)"
trap "$(printf 'rm -rf "%q"' "$TMP")" EXIT

cp "$IPFS_MODFILE" "$TMP/go.mod"
pushd $IPFS_PATH
    ARGS="$($GOCC list -mod=mod -f '-require={{.Path}}@{{.Version}}' -m all | tail -n+2)"
popd

$GOCC mod edit $(echo "$ARGS")
$GOCC mod tidy
