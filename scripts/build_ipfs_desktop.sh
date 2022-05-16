#!/usr/bin/env bash

if [ "$#" -ne "2" ]; then
    echo "usage: $0 <os> <arch>"
    exit 1
fi

OS=$1
ARCH=$2
EXT=""

EBOS=$OS
EBARCH=$ARCH

if [ "$OS" = "darwin" ]; then
    EBOS="macos"
fi

if [ "$OS" = "windows" ]; then
    EXT=".exe"
fi

if [ "$ARCH" = "amd64" ]; then
    EBARCH="x64"
fi

export CSC_IDENTITY_AUTO_DISCOVERY=false

install -D -m 755 bin/ipfs_$OS-$ARCH$EXT third_party/ipfs-desktop/npm-go-ipfs/bin/ipfs$EXT
mkdir -p .cache
docker run --rm -v $(pwd)/third_party/ipfs-desktop/:/project/:Z docker.io/electronuserland/builder:wine sh -c "npm ci && npm run build && npm install && npx electron-builder --$EBOS --$EBARCH"
pushd third_party/ipfs-desktop/dist
    find -maxdepth 1 -name '*aarch64*' -type f -exec sh -c 'mv {} $(echo {} | sed "s/aarch64/arm64/")' \; &> /dev/null
    find -maxdepth 1 -name '*x86_64*' -type f -exec sh -c 'mv {} $(echo {} | sed "s/x86_64/amd64/")' \; &> /dev/null
    find -maxdepth 1 -name '*win*' -type f -exec sh -c 'mv {} $(echo {} | sed "s/win/windows/")' \; &> /dev/null
    cp *.rpm *.deb *.exe *.dmg *.AppImage ../../../bin &> /dev/null
popd
