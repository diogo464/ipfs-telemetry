#!/usr/bin/env bash

source package/package_env.sh
echo sourced
TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

pushd $TMPDIR
mkdir -p usr/lib/systemd/user/
mkdir -p usr/bin/
mkdir -p DEBIAN
popd 

cat package/debian/control | envsubst > "$TMPDIR/DEBIAN/control"
cp bin/ipfs "$TMPDIR/usr/bin/"
cp package/shared/ipfs.service "$TMPDIR/usr/lib/systemd/user"

dpkg-deb --build --root-owner-group "$TMPDIR" bin/ipfs_linux-$PACKAGE_ARCH.deb
