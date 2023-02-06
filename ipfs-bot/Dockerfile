FROM ghcr.io/diogo464/go-ipfs:telemetry-latest
LABEL org.opencontainers.image.source "https://github.com/diogo464/ipfs_telemetry"

ENV PATH=/usr/local/bin:$PATH
ENV PATH=/usr/local/app:$PATH

WORKDIR /usr/src/app

RUN wget https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 -O jq && \
    chmod +x jq && \
    mv jq /usr/local/bin

RUN wget https://github.com/nats-io/natscli/releases/download/v0.0.35/nats-0.0.35-linux-amd64.zip && \
    unzip nats-0.0.35-linux-amd64.zip && \
    mv nats-0.0.35-linux-amd64/nats nats && \
    chmod +x nats && \
    mv nats /usr/local/bin

COPY popular_cids.txt cids.txt
COPY entrypoint.sh .

ENTRYPOINT [ "sh", "./entrypoint.sh" ]