GOCC ?= go
IPFS_BIN ?= ipfs
PROTO_FLAGS := --proto_path=api/ --go_out=./ --go-grpc_out=./ --go-grpc_opt=module=github.com/diogo464/telemetry --go_opt=module=github.com/diogo464/telemetry

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
	protoc $(PROTO_FLAGS) api/*.proto

.PHONY: tools
tools:
	$(GOCC) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOCC) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GOCC) install honnef.co/go/tools/cmd/staticcheck@latest 
	$(GOCC) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
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

