#!/usr/bin/env bash
set -e
cd $(dirname $0)/..

TARGETS=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)

mkdir -p dist
for target in "${TARGETS[@]}"; do
  IFS="/" read -r GOOS GOARCH <<< "$target"
  echo "Building for $GOOS/$GOARCH..."
  OUTPUT="dist/ipfs-$GOOS-$GOARCH" GOOS="$GOOS" GOARCH="$GOARCH" scripts/build-ipfs.sh
done
