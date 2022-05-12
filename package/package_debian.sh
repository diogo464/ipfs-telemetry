#!/usr/bin/env bash

source package/package_env.sh

TMPDIR=$(mktemp -d)
trap "rm -rf $TMPDIR" EXIT

pushd $TMPDIR
mkdir -p usr/lib/systemd/user/
mkdir -p usr/bin/
mkdir -p DEBIAN
popd 

cat package/debian/control | envsubst > "$TMPDIR/DEBIAN/control"
install --mode 644 package/shared/ipfs.service "$TMPDIR/usr/lib/systemd/user"
install --mode 755 bin/ipfs "$TMPDIR/usr/bin/"

dpkg-deb --build --root-owner-group "$TMPDIR" "$PACKAGE_OUTPUT_DIR/ipfs_linux-$PACKAGE_ARCH.deb"
