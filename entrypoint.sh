#!/usr/bin/env sh
IP=$(curl ifconfig.me)
PORT=${PORT:-4001}
API_PORT=${API_PORT:-5001}
GATEWAY_PORT=${GATEWAY_PORT:-8080}
ipfs init --profile server || exit 1
ipfs config Addresses.API "/ip4/0.0.0.0/tcp/$API_PORT" || exit 1
ipfs config Addresses.Gateway "/ip4/127.0.0.1/tcp/$GATEWAY_PORT" || exit 1
ipfs config Routing.Type 'dhtserver' || exit 1
sed -i "s?/\(0.0.0.0\|\:\:\)/\(.*\)/4001\(.*\)?/\1/\2/$PORT\3?g" /root/.ipfs/config
#sed -i "s|/0.0.0.0/\(.*/4001\)|/$IP/\1|" /root/.ipfs/config
ipfs daemon
