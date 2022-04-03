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

build: monitor crawler ipfs

install: ipfs-install

check:
	./scripts/check.sh

.PHONY: proto
proto:
	protoc $(PROTO_FLAGS) api/common.proto
	protoc $(PROTO_FLAGS) api/snapshot.proto
	protoc $(PROTO_FLAGS) api/telemetry.proto
	protoc $(PROTO_FLAGS) api/monitor.proto

generate:
	sqlboiler --wipe psql

setup: tools generate proto

tools:
	$(GOCC) install github.com/volatiletech/sqlboiler/v4@latest
	$(GOCC) install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
	$(GOCC) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOCC) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GOCC) install github.com/BurntSushi/go-sumtype
	$(GOCC) install honnef.co/go/tools/cmd/staticcheck@latest 

tidy:
	./scripts/tidy.sh

database-up:
		@ if podman container exists $(DATABASE_CONTAINER) ; then \
			podman start $(DATABASE_CONTAINER) && sleep 3 ; \
		else \
			podman run --name=$(DATABASE_CONTAINER) --tty --interactive --detach --network=host -e POSTGRES_PASSWORD=$(DATABASE_PASSWORD) $(DATABASE_IMAGE) && sleep 3 ; \
		fi ;

database-down:
	podman rm -f $(DATABASE_CONTAINER)

migrate-up:
	migrate -path $(DATABASE_MIGRATIONS) -database $(DATABASE_URL) up

migrate-down:
	migrate -path $(DATABASE_MIGRATIONS) -database $(DATABASE_URL) down

migrate-version:
	migrate -path $(DATABASE_MIGRATIONS) -database $(DATABASE_URL) version
