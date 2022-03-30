DATABASE_MIGRATIONS := migrations/
DATABASE_PASSWORD := password
DATABASE_URL := postgres://postgres@localhost/postgres?sslmode=disable
DATABASE_CONTAINER := monitor-db
DATABASE_IMAGE := docker.io/library/postgres:14

.PHONY: plugin
plugin:
	GOCC=go1.16 $(MAKE) -B -C plugin/ build
	mkdir -p bin/ && mv plugin/telemetry.so bin/

plugin-install: plugin
	mkdir -p ~/.ipfs/plugins
	cp bin/telemetry.so ~/.ipfs/plugins

monitor:
	go build -o bin/monitor cmd/monitor/*

watch:
	go build -o bin/watch cmd/watch/*

crawler:
	go build -o bin/crawler cmd/crawler/*

build: monitor watch crawler plugin

.PHONY: proto
proto:
	protoc --go_out=./pkg/telemetry/ telemetry.proto
	protoc --go_out=./plugin/ telemetry.proto
	protoc --go_out=./pkg/monitor/ --go-grpc_out=./pkg/monitor/ monitor.proto

generate:
	sqlboiler --wipe psql

setup: tools generate proto

tools:
	go install github.com/volatiletech/sqlboiler/v4@latest
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

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
