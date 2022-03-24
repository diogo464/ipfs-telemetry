FROM docker.io/golang:1.16 as builder

WORKDIR /usr/src/app
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -v -o app

FROM docker.io/debian:sid-slim

EXPOSE 8080

WORKDIR /
COPY --from=builder /usr/src/app/app .
CMD ["./app"]
