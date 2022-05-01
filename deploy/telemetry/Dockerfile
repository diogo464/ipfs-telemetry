FROM docker.io/debian:stable-slim as builder

WORKDIR /usr/src/app
RUN apt -y update && apt -y install wget
RUN wget 'https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=zNX1jd547wMiFH9q&suffix=tar.gz' -O GeoLite2-City.tar.gz 
RUN tar -xf GeoLite2-City.tar.gz && mv GeoLite2-City_*/GeoLite2-City.mmdb .

FROM docker.io/debian:stable-slim

WORKDIR /usr/local/app
COPY --from=builder /usr/src/app/GeoLite2-City.mmdb .
COPY bin/telemetry bin/monitor bin/crawler bin/ipfs bin/link
ENV PATH="/usr/local/app:${PATH}"
