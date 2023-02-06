#!/bin/sh

export NATS_URL="${NATS_ENDPOINT:-nats://nats:4222}"

ipfs init --profile lowpower
ipfs config Addresses.API /ip4/0.0.0.0/tcp/5001
ipfs config Telemetry.MetricsPeriod "8s"
ipfs config Telemetry.WindowDuration "1m"
ipfs config Telemetry.ActiveBufferDuration "10s"

ipfs daemon &
IPFS_PID=$!

CIDS_COUNT=$(wc -l cids.txt | awk '{print $1}')

echo "CIDS: $CIDS_COUNT"
echo "NATS: $NATS_URL"

sleep 10
while true; do
    if kill -0 $IPFS_PID ; then
        DISCOVERY=$(ipfs id | jq '{id: .ID, addresses: .Addresses}')
        nats publish discovery "$DISCOVERY"

        CID_IDX=$(expr $RANDOM % $CIDS_COUNT + 1)
        CID=$(head -n $CID_IDX cids.txt | tail -n 1)
        timeout -s 9 10 ipfs get $CID
        ipfs pin ls | cut -d' ' -f1 | xargs -I{} ipfs pin rm -r {}
        sleep 15
    else
        exit 1
    fi
done