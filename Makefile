GOCC ?= go
GOGOPROTO_FLAGS := -I=. -I=./third_party/ -I=./third_party/github.com/gogo/protobuf/protobuf \
				--gogofast_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types:.

all: telemetry example

.PHONY: example
example:
	$(GOCC) build -o ./bin/basic ./examples/basic

.PHONY: telemetry
telemetry:
	$(GOCC) build -o ./bin/telemetry ./cmd/telemetry

.PHONY: clean
clean:
	rm -rf bin/

.PHONY: tools
tools:
	$(GOCC) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOCC) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GOCC) install honnef.co/go/tools/cmd/staticcheck@latest 
	$(GOCC) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCC) install github.com/gogo/protobuf/protoc-gen-gogofast

.PHONY: proto
proto:
	protoc $(GOGOPROTO_FLAGS) pb/telemetry.proto
	rm -rf pkg/proto
	mv git.d464.sh/uni/telemetry/pb/* pb/
	rm -rf git.d464.sh
