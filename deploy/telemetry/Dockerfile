FROM docker.io/golang:1.18 as builder

RUN apt -y update && apt -y install protobuf-compiler
RUN wget 'https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-City&license_key=zNX1jd547wMiFH9q&suffix=tar.gz' -O GeoLite2-City.tar.gz 
RUN tar -xf GeoLite2-City.tar.gz && mv GeoLite2-City_*/GeoLite2-City.mmdb .

WORKDIR /usr/src/app
COPY third_party third_party
COPY scripts scripts
COPY .git .git
COPY go.mod go.sum Makefile .
RUN make tidy tools
COPY . .
RUN make proto build

FROM docker.io/debian:stable-slim

WORKDIR /usr/local/app
COPY --from=builder /usr/src/app/GeoLite2-City.mmdb .
COPY --from=builder /usr/src/app/bin/* .
ENV PATH="/usr/local/app:${PATH}"
