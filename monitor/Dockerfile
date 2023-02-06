FROM docker.io/golang:1.19.5 as builder

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o monitor ./cmd

FROM docker.io/alpine:3.17.1

LABEL org.opencontainers.image.source "https://github.com/diogo464/ipfs-telemetry"

ENV PATH=/usr/local/app:$PATH
WORKDIR /usr/local/app
COPY --from=builder /usr/src/app/monitor .

ENTRYPOINT [ "monitor" ]