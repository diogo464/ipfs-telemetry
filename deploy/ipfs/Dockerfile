FROM docker.io/golang:1.18-bullseye AS builder
RUN apt update && apt install -y make gcc protobuf-compiler
WORKDIR /app
COPY . ./
RUN CGO_ENABLED=0 make tools proto ipfs

FROM docker.io/alpine:latest
WORKDIR /usr/local/app
COPY --from=builder /app/bin/ipfs .
ENTRYPOINT ["/usr/local/app/ipfs"]
