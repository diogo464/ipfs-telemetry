FROM docker.io/golang:1.16 as builder

WORKDIR /usr/src/ipfs/
COPY ipfs/ .
RUN make build

WORKDIR /usr/src/telemetry/
COPY telemetry/go.* .
RUN go mod edit -dropreplace github.com/ipfs/go-ipfs
RUN go mod download
COPY telemetry/ .
RUN make build IPFS_VERSION=/usr/src/ipfs/

FROM docker.io/debian:sid-slim

WORKDIR /root
COPY --from=builder /usr/src/ipfs/cmd/ipfs/ipfs .
RUN mkdir -p .ipfs/plugins/
COPY --from=builder /usr/src/telemetry/telemetry.so .ipfs/plugins/
RUN ./ipfs init --profile lowpower

CMD ["./ipfs", "daemon"]





