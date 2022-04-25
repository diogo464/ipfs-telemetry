FROM docker.io/fedora:latest

WORKDIR /usr/local/app
COPY bin/* GeoLite2-City.mmdb .
ENV PATH="/usr/local/app:${PATH}"
