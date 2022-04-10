GOCC ?= go
GOFLAGS ?=

DATABASE_MIGRATIONS := migrations/
DATABASE_PASSWORD := password
DATABASE_URL := postgres://postgres@localhost/postgres?sslmode=disable
DATABASE_CONTAINER := monitor-db
DATABASE_IMAGE := docker.io/library/postgres:14

PROTO_FLAGS := --proto_path=api/ --go_out=./ --go-grpc_out=./ --go-grpc_opt=module=git.d464.sh/adc/telemetry --go_opt=module=git.d464.sh/adc/telemetry

ipfs:
	GOCC=$(GOCC) $(MAKE) -B -C third_party/go-ipfs/ build
	mkdir -p bin/ && mv third_party/go-ipfs/cmd/ipfs/ipfs bin/

ipfs-install: ipfs
	cp bin/ipfs ~/.go/bin/

monitor:
	$(GOCC) build -o bin/monitor cmd/monitor/*

crawler:
	$(GOCC) build -o bin/crawler cmd/crawler/*

telemetry:
	$(GOCC) build -o bin/telemetry cmd/telemetry/*

link:
	$(GOCC) build -o bin/link cmd/link/*

build: monitor crawler telemetry ipfs link

install: ipfs-install

check:
	./scripts/check.sh

clean:
	rm -rf bin/ pkg/proto pkg/models

.PHONY: proto
proto:
	protoc $(PROTO_FLAGS) api/common.proto
	protoc $(PROTO_FLAGS) api/snapshot.proto
	protoc $(PROTO_FLAGS) api/telemetry.proto
	protoc $(PROTO_FLAGS) api/monitor.proto
	protoc $(PROTO_FLAGS) api/window.proto
	protoc $(PROTO_FLAGS) api/crawler.proto

generate:
	sqlboiler --wipe psql

setup: tools generate proto

tools:
	$(GOCC) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOCC) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GOCC) install honnef.co/go/tools/cmd/staticcheck@latest 
	$(GOCC) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	mkdir -p bin/ && cd third_party/go-sumtype && $(GOCC) build -o ../../bin

tidy:
	./scripts/tidy.sh

database-up:
	./scripts/database.sh up

database-down:
	./scripts/database.sh down
