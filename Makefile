GOCC ?= go
GOFLAGS ?=

DATABASE_MIGRATIONS := migrations/
DATABASE_PASSWORD := password
DATABASE_URL := postgres://postgres@localhost/postgres?sslmode=disable
DATABASE_CONTAINER := monitor-db
DATABASE_IMAGE := docker.io/library/postgres:14

ipfs:
	GOCC=$(GOCC) $(MAKE) -B -C third_party/go-ipfs/ build
	mkdir -p bin/ && mv third_party/go-ipfs/cmd/ipfs/ipfs bin/

ipfs-install: ipfs
	cp bin/ipfs ~/.go/bin/

monitor:
	$(GOCC) build -o bin/monitor cmd/monitor/*

watch:
	$(GOCC) build -o bin/watch cmd/watch/*

crawler:
	$(GOCC) build -o bin/crawler cmd/crawler/*

test:
	$(GOCC) build -o bin/test cmd/test/*

build: monitor watch crawler test ipfs

install: ipfs-install

.PHONY: proto
proto:
	protoc --go_out=./pkg/telemetry/ --go-grpc_out=./pkg/telemetry/ telemetry.proto
	protoc --go_out=./pkg/monitor/ --go-grpc_out=./pkg/monitor/ monitor.proto

generate:
	sqlboiler --wipe psql

setup: tools generate proto

tools:
	$(GOCC) install github.com/volatiletech/sqlboiler/v4@latest
	$(GOCC) install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
	$(GOCC) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOCC) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

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
