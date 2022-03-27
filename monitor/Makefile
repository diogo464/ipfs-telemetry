DATABASE_MIGRATIONS := migrations/
DATABASE_PASSWORD := password
DATABASE_URL := postgres://postgres@localhost/postgres?sslmode=disable
DATABASE_CONTAINER := monitor-db
DATABASE_IMAGE := docker.io/library/postgres:14

setup: tools database-up migrate-up generate

tools:
	go install github.com/volatiletech/sqlboiler/v4@latest
	go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

generate:
	sqlboiler --wipe psql

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
