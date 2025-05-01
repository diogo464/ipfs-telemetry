#!/usr/bin/env -S bash -x

FILE="GeoLite2-City.tar.gz"
DIR="$(mktemp -d)"
trap 'rm -rf -- $DIR' EXIT

pushd "$DIR"
    wget 'https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=zNX1jd547wMiFH9q&suffix=tar.gz' -O $FILE || exit 1
    tar -xf $FILE || exit 1
    wget 'https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=zNX1jd547wMiFH9q&suffix=tar.gz' -O $FILE || exit 1
    tar -xf $FILE || exit 1
popd
mv "$DIR"/GeoLite2-City_*/GeoLite2-City.mmdb .
mv "$DIR"/GeoLite2-ASN_*/GeoLite2-ASN.mmdb .
