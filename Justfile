_default:
    just --list

# setup remotes used by the sourced projects
setup-remotes:
    git remote add kubo https://github.com/ipfs/kubo
    git remote add boxo https://github.com/ipfs/boxo
    git remote add kad https://github.com/libp2p/go-libp2p-kad-dht
    git remote add libp2p https://github.com/libp2p/go-libp2p

tidy:
    cd boxo && go mod tidy
    cd kad && go mod tidy
    cd kubo && go mod tidy
    cd libp2p && go mod tidy
    cd monitor && go mod tidy
    cd crawler && go mod tidy
    cd telemetry && go mod tidy

# build the ipfs binary
build-ipfs:
    ./scripts/build-ipfs.sh

# build the telemetry binary
build-telemetry:
    ./scripts/build-telemetry.sh

# build the backend binary
build-backend:
    ./scripts/build-backend.sh

# build all binaries
build: build-ipfs build-telemetry build-backend

build-backend-container: build-backend
    ./scripts/build-backend-container.sh

build-ipfs-container:
    ./scripts/build-ipfs-container.sh

build-grafana-container:
    ./scripts/build-grafana-container.sh

build-container: build-ipfs-container build-backend-container build-grafana-container

# fetch the nats cli
fetch-nats:
    ./scripts/fetch-nats-cli.sh

# start the nats container
nats:
    ./scripts/podman-nats.sh

# start the victoria metrics container
vm:
    ./scripts/podman-vm.sh

# start the postgres container
pg:
    ./scripts/podman-pg.sh

grafana:
    ./scripts/podman-grafana.sh

monitor: build-backend
    ./scripts/monitor.sh

crawler: build-backend
    ./scripts/crawler.sh

exporter-vm: build-backend
    ./scripts/exporter-vm.sh

exporter-pg: build-backend
    ./scripts/exporter-pg.sh

# show logs for a given container
logs name:
    podman logs -f {{name}}

sync:
    rsync -avzp --exclude data --exclude .git --exclude dist --exclude bin . ipfs:./

hsync:
    rsync -avzp --exclude data --exclude .git --exclude dist --exclude bin . hipfs:./
