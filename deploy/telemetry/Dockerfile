FROM docker.io/fedora:latest

WORKDIR /usr/local/app
COPY GeoLite2-City.mmdb .
COPY bin/* .
ENV PATH="/usr/local/app:${PATH}"
