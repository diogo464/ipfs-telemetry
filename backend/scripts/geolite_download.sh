#!/usr/bin/sh

FILE="GeoLite2-City.tar.gz"
DIR="$(mktemp -d)"
CDIR="$(pwd)"
trap 'rm -rf -- $DIR' EXIT

cd "$DIR"
    wget 'https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=zNX1jd547wMiFH9q&suffix=tar.gz' -O $FILE || exit 1
    tar -xf $FILE || exit 1
cd "$CDIR"
mv "$DIR"/GeoLite2-City_*/GeoLite2-City.mmdb .
