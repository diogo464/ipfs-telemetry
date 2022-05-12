#!/usr/bin/env bash

source package/package_env.sh

TMPDIR=$(mktemp -d)
#trap "rm -rf $TMPDIR" EXIT

export PACKAGE_ARCHLINUX_SHA256_IPFS_SERVICE=$(sha256sum package/shared/ipfs.service | cut -d' ' -f1)
export PACKAGE_ARCHLINUX_SHA256_IPFS=$(sha256sum bin/ipfs | cut -d' ' -f1)
cat package/archlinux/PKGBUILD | envsubst '${PACKAGE_VERSION},${PACKAGE_ARCH},${PACKAGE_ARCHLINUX_SHA256_IPFS},${PACKAGE_ARCHLINUX_SHA256_IPFS_SERVICE}' > $TMPDIR/PKGBUILD
cp package/shared/ipfs.service $TMPDIR
cp bin/ipfs $TMPDIR
    
pushd $TMPDIR
    makepkg
popd

cp $TMPDIR/*.pkg.tar.zst "$PACKAGE_OUTPUT_DIR/ipfs_linux-$PACKAGE_ARCH.pkg.tar.zst"
