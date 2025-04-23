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
    cd telemetry && go mod tidy

# build the ipfs binary
build-ipfs:
    ./scripts/build-ipfs.sh

# build the telemetry binary
build-telemetry:
    ./scripts/build-telemetry.sh

# build the monitor binary
build-monitor:
    ./scripts/build-monitor.sh

# build all binaries
build: build-ipfs build-telemetry build-monitor

# fetch the nats cli
fetch-nats:
    ./scripts/fetch-nats-cli.sh

# start the nats container
nats:
    ./scripts/podman-nats.sh

# start the victoria metrics container
vm:
    ./scripts/podman-vm.sh

monitor: build-monitor
    MONITOR_COLLECT_INTERVAL=5s bin/monitor

exporter-vm:
    cd backend && poetry run python -m jobs.export-vm

webapi:
    cd backend && poetry run python -m webapi

# show logs for a given container
logs name:
    podman logs -f {{name}}
