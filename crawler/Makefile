.DEFAULT_GOAL := build

GOCC ?= CGO_ENABLED=0 go

.PHONY: check
check:
	go vet ./...
	staticcheck ./...
	golangci-lint run ./...

.PHONY: tools
tools:
	$(GOCC) install honnef.co/go/tools/cmd/staticcheck@latest 
	$(GOCC) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCC) install github.com/gogo/protobuf/protoc-gen-gogofast

.PHONY: build
build:
	go build -o crawler ./cmd
