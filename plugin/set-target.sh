#!/bin/bash

GOCC="${GOCC:-go}"

set -eo pipefail

GOPATH="$($GOCC env GOPATH)"
IPFS_PATH="$(pwd)/../../ipfs"

MODFILE="$IPFS_PATH/go.mod"
$GOCC mod edit -replace "github.com/ipfs/go-ipfs=$IPFS_PATH"
$GOCC mod edit -replace "git.d464.sh/adc/rle=$(pwd)/../../rle"

TMP="$(mktemp -d)"
trap "$(printf 'rm -rf "%q"' "$TMP")" EXIT

(
    cd "$TMP"
    cp "$MODFILE" "go.mod"
    go list -mod=mod -f '-require={{.Path}}@{{.Version}}{{if .Replace}} -replace={{.Path}}@{{.Version}}={{.Replace}}{{end}}' -m all | tail -n+2  > args
)

$GOCC mod edit $(cat "$TMP/args")
$GOCC mod tidy
