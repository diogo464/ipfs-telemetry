_default:
    just --list

# setup remotes used by the sourced projects
setup-remotes:
    git remote add kubo https://github.com/ipfs/kubo
    git remote add boxo https://github.com/ipfs/boxo
    git remote add kad https://github.com/libp2p/go-libp2p-kad-dht
    git remote add libp2p https://github.com/libp2p/go-libp2p

# build the ipfs binary
build-ipfs:
    ./scripts/build-ipfs.sh

# build the telemetry binary
build-telemetry:
    ./scripts/build-telemetry.sh

# build all binaries
build: build-ipfs build-telemetry

# fetch the nats cli
fetch-nats:
    ./scripts/fetch-nats-cli.sh

# start the nats container
container-nats:
    ./scripts/podman-nats.sh

# start the victoria metrics container
container-vm:
    ./scripts/podman-vm.sh
