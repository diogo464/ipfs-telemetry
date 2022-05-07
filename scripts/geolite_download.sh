#!/usr/bin/env bash

FILE="GeoLite2-City.tar.gz"
DIR="$(mktemp -d)"
trap 'rm -rf -- $DIR' EXIT

pushd "$DIR" || exit
    wget 'https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=zNX1jd547wMiFH9q&suffix=tar.gz' -O $FILE || exit 1
    tar -xf $FILE || exit 1
popd || exit
mv "$DIR"/GeoLite2-City_*/GeoLite2-City.mmdb .
