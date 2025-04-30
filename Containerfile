FROM docker.io/alpine:latest
RUN apk add curl
COPY ./bin/ipfs /usr/bin/ipfs
COPY ./entrypoint.sh /usr/bin/entrypoint.sh
ENTRYPOINT ["/usr/bin/entrypoint.sh"]
