default:
    just --list

setup-remotes:
    git remote add kubo https://github.com/ipfs/kubo
    git remote add boxo https://github.com/ipfs/boxo
    git remote add kad https://github.com/libp2p/go-libp2p-kad-dht
    git remote add libp2p https://github.com/libp2p/go-libp2p
