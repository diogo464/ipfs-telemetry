FROM docker.io/golang:1.18 as builder

WORKDIR /usr/src/app
COPY . .
RUN make tidy && make crawler

FROM docker.io/debian:stable-slim

WORKDIR /usr/local/app
COPY --from=builder /usr/src/app/bin/crawler .
RUN ./crawler
