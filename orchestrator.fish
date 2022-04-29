#!/bin/fish

set IPS (./deploy/probing/azure.fish ip)
set PROBES "127.0.0.1:4640"

for IP in $IPS
    set PROBES "$PROBES,$IP:4640"
end

echo $PROBES
bin/orchestrator --num-cids 1 --influxdb-address http://127.0.0.1:8086 --influxdb-token token --influxdb-org adc --influxdb-bucket probing --probes $PROBES
