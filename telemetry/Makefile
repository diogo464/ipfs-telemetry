GOCC ?= go
GOGOPROTO_FLAGS := -I=. -I=./third_party/opentelemetry-proto/ --go_out=. --go-grpc_out=.

all: telemetry example

.PHONY: example
example:
	$(GOCC) build -o ./bin/basic ./examples/basic

.PHONY: telemetry
telemetry:
	$(GOCC) build -o ./bin/telemetry ./cmd/telemetry

.PHONY: check
check:
	$(GOCC) vet ./...
	staticcheck ./...
	golangci-lint run

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: tools
tools:
	$(GOCC) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOCC) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GOCC) install honnef.co/go/tools/cmd/staticcheck@latest 
	$(GOCC) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: proto
proto:
	protoc $(GOGOPROTO_FLAGS) internal/pb/telemetry.proto
	protoc $(GOGOPROTO_FLAGS) crawler/pb/crawler.proto
