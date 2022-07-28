GOCC ?= go
IPFS_BIN ?= ipfs
GOGOPROTO_FLAGS := -I=api/ -I=. -I=./third_party/ -I=./third_party/github.com/gogo/protobuf/protobuf \
				--gogofast_out=plugins=grpc,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/struct.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/wrappers.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/empty.proto=github.com/gogo/protobuf/types:.

.PHONY: setup
setup: tools proto geolite

.PHONY: check
check:
	./scripts/check.sh

.PHONY: tidy
tidy:
	./scripts/tidy.sh

GeoLite2-City.mmdb:
	./scripts/geolite_download.sh

.PHONY: geolite
geolite: GeoLite2-City.mmdb

.PHONY: proto
proto:
	protoc $(GOGOPROTO_FLAGS) api/common.proto
	protoc $(GOGOPROTO_FLAGS) api/crawler.proto
	protoc $(GOGOPROTO_FLAGS) api/datapoint.proto
	protoc $(GOGOPROTO_FLAGS) api/monitor.proto
	protoc $(GOGOPROTO_FLAGS) api/probe.proto
	rm -rf pkg/proto
	mv github.com/diogo464/ipfs_telemetry/pkg/proto pkg/
	rm -rf github.com

.PHONY: tools
tools:
	$(GOCC) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOCC) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GOCC) install honnef.co/go/tools/cmd/staticcheck@latest 
	$(GOCC) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOCC) install github.com/gogo/protobuf/protoc-gen-gogofast
	mkdir -p bin/ && cd third_party/go-sumtype && $(GOCC) build -o ../../bin

.PHONY: clean
clean:
	rm -rf bin/ pkg/proto pkg/models

.PHONY: build
build: build-telemetry build-probing

.PHONY: build-telemetry
build-telemetry: monitor crawler telemetry ipfs link walker

.PHONY: build-probing
build-probing: orchestrator probe

.PHONY: ipfs
ipfs:
	$(MAKE) -B -C third_party/go-ipfs/ build
	mkdir -p bin/ && mv third_party/go-ipfs/cmd/ipfs/ipfs bin/$(IPFS_BIN)

.PHONY: monitor
monitor:
	$(GOCC) build -o bin/monitor cmd/monitor/*

.PHONY: crawler
crawler:
	$(GOCC) build -o bin/crawler cmd/crawler/*

.PHONY: telemetry
telemetry:
	$(GOCC) build -o bin/telemetry cmd/telemetry/*

.PHONY: link
link:
	$(GOCC) build -o bin/link cmd/link/*

.PHONY: orchestrator
orchestrator:
	$(GOCC) build -o bin/orchestrator cmd/orchestrator/*

.PHONY: probe
probe:
	$(GOCC) build -o bin/probe cmd/probe/*

.PHONY: walker
walker:
	$(GOCC) build -o bin/walker cmd/walker/*

.PHONY: docker-telemetry
docker-telemetry: build-telemetry
	podman build -t ghcr.io/diogo464/telemetry:latest -f deploy/telemetry/Dockerfile .

.PHONY: docker-probing
docker-probing: build-probing
	podman build -t ghcr.io/diogo464/probing:latest -f deploy/probing/Dockerfile .

.PHONY: docker
docker: docker-telemetry docker-probing

