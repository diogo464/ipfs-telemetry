FROM docker.io/alpine:latest
WORKDIR /usr/local/app
COPY ./bin/backend /usr/bin/backend
COPY GeoLite2-ASN.mmdb .
COPY GeoLite2-City.mmdb .
ENTRYPOINT ["/usr/bin/backend"]
